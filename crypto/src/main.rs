use std::io::{self, Read};

use serde::Deserialize;
use serde_json::{Value, json};
use vodozemac::{
    Curve25519PublicKey,
    olm::{Account, AccountPickle, OlmMessage, Session, SessionConfig, SessionPickle},
};

#[derive(Deserialize)]
#[serde(tag = "op", rename_all = "snake_case", deny_unknown_fields)]
enum Request {
    Create {
        one_time_keys: usize,
    },
    Replenish {
        account: Value,
        one_time_keys: usize,
    },
    Encrypt {
        account: Value,
        session: Option<Value>,
        peer_identity: Option<String>,
        peer_one_time_key: Option<String>,
        plaintext: String,
    },
    Decrypt {
        account: Value,
        session: Option<Value>,
        sender_identity: String,
        message: OlmMessage,
    },
}

fn publish_keys(account: &mut Account, count: usize) -> Result<Vec<Value>, &'static str> {
    if !(1..=20).contains(&count)
        || account.stored_one_time_key_count() + count > account.max_number_of_one_time_keys()
    {
        return Err("invalid key count");
    }
    let existing: std::collections::HashSet<String> = account
        .one_time_keys()
        .values()
        .map(ToString::to_string)
        .collect();
    account.generate_one_time_keys(count);
    let identity = account.identity_keys();
    let keys: Vec<Value> = account
        .one_time_keys()
        .values()
        .filter(|key| !existing.contains(&key.to_string()))
        .map(|key| {
            let payload = format!("KARATTY-OTK-v1\n{}\n{}", identity.curve25519, key);
            json!({"key": key.to_base64(), "signature": account.sign(payload).to_base64()})
        })
        .collect();
    account.mark_keys_as_published();
    Ok(keys)
}

fn handle(request: Request) -> Result<Value, &'static str> {
    match request {
        Request::Create { one_time_keys } => {
            let mut account = Account::new();
            let keys = publish_keys(&mut account, one_time_keys)?;
            let identity = account.identity_keys();
            Ok(
                json!({"account": account.pickle(), "identity": {"ed25519": identity.ed25519.to_base64(), "curve25519": identity.curve25519.to_base64()}, "one_time_keys": keys}),
            )
        }
        Request::Replenish {
            account,
            one_time_keys,
        } => {
            let account: AccountPickle =
                serde_json::from_value(account).map_err(|_| "invalid account")?;
            let mut account = Account::from_pickle(account);
            let keys = publish_keys(&mut account, one_time_keys)?;
            Ok(json!({"account": account.pickle(), "one_time_keys": keys}))
        }
        Request::Encrypt {
            account,
            session,
            peer_identity,
            peer_one_time_key,
            plaintext,
        } => {
            if plaintext.is_empty() || plaintext.len() > 6144 {
                return Err("invalid plaintext length");
            }
            let account: AccountPickle =
                serde_json::from_value(account).map_err(|_| "invalid account")?;
            let account = Account::from_pickle(account);
            let mut session = if let Some(pickle) = session {
                let pickle: SessionPickle =
                    serde_json::from_value(pickle).map_err(|_| "invalid session")?;
                Session::from_pickle(pickle)
            } else {
                let identity = Curve25519PublicKey::from_base64(
                    &peer_identity.ok_or("missing peer identity")?,
                )
                .map_err(|_| "invalid peer identity")?;
                let one_time_key = Curve25519PublicKey::from_base64(
                    &peer_one_time_key.ok_or("missing peer one-time key")?,
                )
                .map_err(|_| "invalid peer one-time key")?;
                account
                    .create_outbound_session(SessionConfig::version_1(), identity, one_time_key)
                    .map_err(|_| "session creation failed")?
            };
            let message = session
                .encrypt(plaintext.as_bytes())
                .map_err(|_| "encryption failed")?;
            Ok(
                json!({"account": account.pickle(), "session": session.pickle(), "session_id": session.session_id(), "message": message}),
            )
        }
        Request::Decrypt {
            account,
            session,
            sender_identity,
            message,
        } => {
            let account: AccountPickle =
                serde_json::from_value(account).map_err(|_| "invalid account")?;
            let mut account = Account::from_pickle(account);
            let sender_identity = Curve25519PublicKey::from_base64(&sender_identity)
                .map_err(|_| "invalid sender identity")?;
            let (session, plaintext) = if let Some(pickle) = session {
                let pickle: SessionPickle =
                    serde_json::from_value(pickle).map_err(|_| "invalid session")?;
                let mut session = Session::from_pickle(pickle);
                let plaintext = session.decrypt(&message).map_err(|_| "decryption failed")?;
                (session, plaintext)
            } else if let OlmMessage::PreKey(ref pre_key) = message {
                let result = account
                    .create_inbound_session(SessionConfig::version_1(), sender_identity, pre_key)
                    .map_err(|_| "inbound session creation failed")?;
                (result.session, result.plaintext)
            } else {
                return Err("unknown session");
            };
            let plaintext = String::from_utf8(plaintext).map_err(|_| "invalid plaintext")?;
            Ok(
                json!({"account": account.pickle(), "session": session.pickle(), "session_id": session.session_id(), "plaintext": plaintext}),
            )
        }
    }
}

fn main() {
    let mut input = Vec::new();
    let result = io::stdin()
        .take(1_048_577)
        .read_to_end(&mut input)
        .map_err(|_| "read failed")
        .and_then(|_| {
            if input.len() > 1_048_576 {
                return Err("request too large");
            }
            let request = serde_json::from_slice(&input).map_err(|_| "invalid request")?;
            handle(request)
        });
    match result {
        Ok(response) => println!("{response}"),
        Err(reason) => {
            eprintln!("karatty-crypto: {reason}");
            std::process::exit(1);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn two_accounts_exchange_and_reject_tampering() {
        let alice = handle(Request::Create { one_time_keys: 1 }).unwrap();
        let bob = handle(Request::Create { one_time_keys: 2 }).unwrap();
        let first = handle(Request::Encrypt {
            account: alice["account"].clone(),
            session: None,
            peer_identity: Some(
                serde_json::from_value(bob["identity"]["curve25519"].clone()).unwrap(),
            ),
            peer_one_time_key: Some(
                serde_json::from_value(bob["one_time_keys"][0]["key"].clone()).unwrap(),
            ),
            plaintext: "hola".into(),
        })
        .unwrap();
        let received = handle(Request::Decrypt {
            account: bob["account"].clone(),
            session: None,
            sender_identity: serde_json::from_value(alice["identity"]["curve25519"].clone())
                .unwrap(),
            message: serde_json::from_value(first["message"].clone()).unwrap(),
        })
        .unwrap();
        assert_eq!(received["plaintext"], "hola");
        let second = handle(Request::Encrypt {
            account: first["account"].clone(),
            session: Some(first["session"].clone()),
            peer_identity: None,
            peer_one_time_key: None,
            plaintext: "segundo".into(),
        })
        .unwrap();
        let third = handle(Request::Encrypt {
            account: second["account"].clone(),
            session: Some(second["session"].clone()),
            peer_identity: None,
            peer_one_time_key: None,
            plaintext: "tercero".into(),
        })
        .unwrap();
        let out_of_order = handle(Request::Decrypt {
            account: received["account"].clone(),
            session: Some(received["session"].clone()),
            sender_identity: serde_json::from_value(alice["identity"]["curve25519"].clone())
                .unwrap(),
            message: serde_json::from_value(third["message"].clone()).unwrap(),
        })
        .unwrap();
        assert_eq!(out_of_order["plaintext"], "tercero");
        let skipped = handle(Request::Decrypt {
            account: out_of_order["account"].clone(),
            session: Some(out_of_order["session"].clone()),
            sender_identity: serde_json::from_value(alice["identity"]["curve25519"].clone())
                .unwrap(),
            message: serde_json::from_value(second["message"].clone()).unwrap(),
        })
        .unwrap();
        assert_eq!(skipped["plaintext"], "segundo");
        let replenished = handle(Request::Replenish {
            account: received["account"].clone(),
            one_time_keys: 1,
        })
        .unwrap();
        assert_eq!(replenished["one_time_keys"].as_array().unwrap().len(), 1);
        assert_ne!(
            replenished["one_time_keys"][0]["key"],
            bob["one_time_keys"][0]["key"]
        );

        let reply = handle(Request::Encrypt {
            account: received["account"].clone(),
            session: Some(received["session"].clone()),
            peer_identity: None,
            peer_one_time_key: None,
            plaintext: "respuesta".into(),
        })
        .unwrap();
        let decrypted = handle(Request::Decrypt {
            account: first["account"].clone(),
            session: Some(first["session"].clone()),
            sender_identity: serde_json::from_value(bob["identity"]["curve25519"].clone()).unwrap(),
            message: serde_json::from_value(reply["message"].clone()).unwrap(),
        })
        .unwrap();
        assert_eq!(decrypted["plaintext"], "respuesta");

        let message: OlmMessage = serde_json::from_value(reply["message"].clone()).unwrap();
        let (kind, mut bytes) = message.to_parts();
        let last = bytes.len() - 1;
        bytes[last] ^= 1;
        if let Ok(damaged) = OlmMessage::from_parts(kind, &bytes) {
            assert!(
                handle(Request::Decrypt {
                    account: first["account"].clone(),
                    session: Some(first["session"].clone()),
                    sender_identity: serde_json::from_value(bob["identity"]["curve25519"].clone())
                        .unwrap(),
                    message: damaged,
                })
                .is_err()
            );
        }
        assert!(
            handle(Request::Decrypt {
                account: decrypted["account"].clone(),
                session: Some(decrypted["session"].clone()),
                sender_identity: serde_json::from_value(bob["identity"]["curve25519"].clone())
                    .unwrap(),
                message,
            })
            .is_err()
        );
    }
}
