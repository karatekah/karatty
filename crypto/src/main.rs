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
    Encrypt {
        account: AccountPickle,
        session: Option<SessionPickle>,
        peer_identity: Option<Curve25519PublicKey>,
        peer_one_time_key: Option<Curve25519PublicKey>,
        plaintext: String,
    },
    Decrypt {
        account: AccountPickle,
        session: Option<SessionPickle>,
        sender_identity: Curve25519PublicKey,
        message: OlmMessage,
    },
}

fn handle(request: Request) -> Result<Value, &'static str> {
    match request {
        Request::Create { one_time_keys } => {
            if !(1..=20).contains(&one_time_keys) {
                return Err("invalid key count");
            }
            let mut account = Account::new();
            account.generate_one_time_keys(one_time_keys);
            let identity = account.identity_keys();
            let keys: Vec<Value> = account
                .one_time_keys()
                .values()
                .map(|key| {
                    let payload = format!("KARATTY-OTK-v1\n{}\n{}", identity.curve25519, key);
                    json!({"key": key, "signature": account.sign(payload).to_base64()})
                })
                .collect();
            account.mark_keys_as_published();
            Ok(json!({"account": account.pickle(), "identity": identity, "one_time_keys": keys}))
        }
        Request::Encrypt {
            account,
            session,
            peer_identity,
            peer_one_time_key,
            plaintext,
        } => {
            if plaintext.is_empty() || plaintext.len() > 4096 {
                return Err("invalid plaintext length");
            }
            let account = Account::from_pickle(account);
            let mut session = if let Some(pickle) = session {
                Session::from_pickle(pickle)
            } else {
                let identity = peer_identity.ok_or("missing peer identity")?;
                let one_time_key = peer_one_time_key.ok_or("missing peer one-time key")?;
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
            let mut account = Account::from_pickle(account);
            let (session, plaintext) = if let Some(pickle) = session {
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
        let bob = handle(Request::Create { one_time_keys: 1 }).unwrap();
        let first = handle(Request::Encrypt {
            account: serde_json::from_value(alice["account"].clone()).unwrap(),
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
            account: serde_json::from_value(bob["account"].clone()).unwrap(),
            session: None,
            sender_identity: serde_json::from_value(alice["identity"]["curve25519"].clone())
                .unwrap(),
            message: serde_json::from_value(first["message"].clone()).unwrap(),
        })
        .unwrap();
        assert_eq!(received["plaintext"], "hola");

        let reply = handle(Request::Encrypt {
            account: serde_json::from_value(received["account"].clone()).unwrap(),
            session: Some(serde_json::from_value(received["session"].clone()).unwrap()),
            peer_identity: None,
            peer_one_time_key: None,
            plaintext: "respuesta".into(),
        })
        .unwrap();
        let decrypted = handle(Request::Decrypt {
            account: serde_json::from_value(first["account"].clone()).unwrap(),
            session: Some(serde_json::from_value(first["session"].clone()).unwrap()),
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
                    account: serde_json::from_value(first["account"].clone()).unwrap(),
                    session: Some(serde_json::from_value(first["session"].clone()).unwrap()),
                    sender_identity: serde_json::from_value(bob["identity"]["curve25519"].clone())
                        .unwrap(),
                    message: damaged,
                })
                .is_err()
            );
        }
        assert!(
            handle(Request::Decrypt {
                account: serde_json::from_value(decrypted["account"].clone()).unwrap(),
                session: Some(serde_json::from_value(decrypted["session"].clone()).unwrap()),
                sender_identity: serde_json::from_value(bob["identity"]["curve25519"].clone())
                    .unwrap(),
                message,
            })
            .is_err()
        );
    }
}
