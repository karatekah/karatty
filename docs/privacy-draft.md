# Política de privacidad de Karatty — borrador para revisión

**No publicar todavía.** Faltan la identidad y el correo público del titular, el dominio, las ubicaciones reales de los proveedores, la comprobación de la marca y la validación legal y operativa. Fecha de entrada en vigor: `[PENDIENTE]`.

## Responsable y contacto

Karatty es operado por `[NOMBRE LEGAL DEL TITULAR]`, persona natural domiciliada en Perú. Para consultas de privacidad, acceso, rectificación, cancelación u oposición, escribe a `[CORREO PÚBLICO DE PRIVACIDAD]`. Para soporte: `[CORREO PÚBLICO DE SOPORTE]`.

## Datos que tratamos y finalidad

- **Cuenta y acceso:** seudónimo, clave pública SSH, huella de esa clave, fechas de creación y acceso, hash de las sesiones y hash del código de recuperación. Se usan para crear la cuenta, autenticarte, recuperar el acceso y prevenir abusos. No pedimos correo, nombre real ni clave privada SSH; tampoco almacenamos el código de recuperación en claro.
- **Claves de E2EE (experimental):** identidad pública Olm, firma con la clave SSH, claves públicas de un solo uso firmadas y marca de consumo. Se guardan para distribuir claves solo a interlocutores de un directo aceptado y evitar reutilizar las ya consumidas. El servidor no recibe la identidad privada ni el estado del ratchet; estas claves públicas se eliminan al borrar la cuenta o recuperar el acceso con otra clave SSH. El cliente local usa estas claves para cifrar los directos del modo público de ensayo.
- **Archivo local cifrado:** el historial legible, los envíos pendientes y el estado del ratchet se guardan cifrados en el equipo, con una frase secreta local independiente de la clave SSH. El cliente la solicita al abrir el archivo y elimina entradas vencidas al volver a usarlo. Perderla impide abrir el archivo; el código de recuperación de cuenta no la reemplaza.
- **Mensajería:** en el modo público de ensayo, el servidor guarda sobres cifrados, remitente, conversación, identificador de reintento, secuencia y fecha; solicitudes y bloqueos de chats directos. El modo por invitación anterior conserva cuerpos legibles y canales, por lo que no formará parte del servicio público. Los metadatos permiten entregar mensajes y controlar consentimiento y bloqueo.
- **Operación:** contadores agregados de peticiones y fallos y límites temporales de tasa en memoria. La aplicación no registra el cuerpo de mensajes, códigos ni tokens en sus logs operativos. El proveedor de infraestructura puede generar registros propios conforme a su servicio.

El seudónimo no garantiza anonimato absoluto: la clave pública, la relación entre interlocutores, las fechas y la información que compartas pueden identificarte. Los miembros de una conversación pueden leer los mensajes enviados a ella. En el modo público de ensayo, el cliente oficial cifra el contenido antes de enviarlo; el operador conserva los metadatos citados. El modo legado sigue exponiendo texto legible al servidor. **Este prototipo no es todavía un servicio público.** HTTPS y WSS protegen el tránsito.

## Proveedores y transferencias

El servidor y PostgreSQL se alojarán en `[PROVEEDOR, REGIÓN Y PAÍS]`. Los respaldos cifrados se guardarán en `[PROVEEDOR, REGIÓN Y PAÍS]`. El código de recuperación se muestra una sola vez en el cliente y no se envía por correo. Antes del lanzamiento se completará la información sobre ubicaciones y flujos internacionales de datos.

No vendemos datos personales ni usamos el contenido de mensajes para publicidad o entrenamiento de modelos. `[VERIFICAR CON LOS CONTRATOS Y CONFIGURACIONES REALES DE LOS PROVEEDORES]`.

## Conservación y borrado

- Los mensajes dejan de estar disponibles en el historial a las **720 horas (30 días)** de su creación. La purga física de la base activa se ejecuta al arrancar y cada hora en modo público; puede haber un breve desfase operativo tras el límite de acceso.
- Las solicitudes de chat directo vencen a los **7 días**. Los desafíos de autenticación y recuperación vencen a los **2 minutos**. Las sesiones temporales vencen a las **8 horas**; sus registros caducados se limpiarán automáticamente antes de publicar esta política.
- El usuario puede borrar su cuenta desde el cliente. La operación borra sus claves SSH y públicas de E2EE, sesiones, membresías y mensajes que escribió de la base activa. Los mensajes escritos por otras personas se conservan hasta su propio vencimiento; los chats directos afectados quedan en solo lectura para ellas.
- Los respaldos completos se cifran antes de subirlos. La operación propuesta los elimina a los **3 días** y configura el mismo plazo en el proveedor. Un mensaje ya vencido o una cuenta borrada pueden permanecer en un respaldo por un período adicional, con un máximo público previsto de **7 días**. Este máximo debe verificarse contra el borrado real del proveedor antes de publicar esta política.

## Tus opciones y derechos

Puedes rechazar solicitudes de chat directo, bloquear usuarios y borrar tu cuenta. Para ejercer derechos de acceso, rectificación, cancelación u oposición sobre tus datos, escribe al correo de privacidad indicado arriba. Podemos pedir una prueba razonable de que controlas la cuenta antes de responder. La [ANPD explica los derechos ARCO](https://www.gob.pe/9270-que-son-los-derechos-arco) y las vías de reclamo disponibles en Perú.

## Seguridad y cambios

Protegemos el acceso con claves SSH y un código de recuperación que se muestra una sola vez. Conserva tu clave privada y ese código fuera del alcance de terceros. Si pierdes ambos, el flujo automático no puede recuperar tu cuenta. Publicaremos los cambios materiales de esta política en `[URL PÚBLICA]` antes de que entren en vigor. La recuperación mediante código reemplaza la clave SSH; antes del lanzamiento se deberá aclarar cómo afecta esto a la lectura de mensajes E2EE anteriores.

## Revisión previa a publicación

La [ANPD indica que los titulares de bancos de datos personales, incluidas personas naturales, deben inscribirlos](https://www.gob.pe/institucion/anpd/pages/8060-inscribir-banco-de-datos-en-el-registro-nacional-de-proteccion-de-datos-personales). La [comunicación de flujos transfronterizos](https://www.gob.pe/9253-inscribir-flujo-transfronterizo-de-datos-personales) también debe revisarse al confirmar las ubicaciones de infraestructura. Estas gestiones corresponden al titular y requieren sus datos personales; este borrador no afirma que ya se hayan completado.
