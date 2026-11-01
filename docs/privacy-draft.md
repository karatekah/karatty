# Política de privacidad de Karatty — borrador para revisión

**No publicar todavía.** Faltan la identidad y el correo público del titular, el dominio, las ubicaciones reales de los proveedores, la comprobación de la marca y la validación legal y operativa. Fecha de entrada en vigor: `[PENDIENTE]`.

## Responsable y contacto

Karatty es operado por `[NOMBRE LEGAL DEL TITULAR]`, persona natural domiciliada en Perú. Para consultas de privacidad, acceso, rectificación, cancelación u oposición, escribe a `[CORREO PÚBLICO DE PRIVACIDAD]`. Para soporte: `[CORREO PÚBLICO DE SOPORTE]`.

## Datos que tratamos y finalidad

- **Cuenta y acceso:** nombre de usuario, correo electrónico verificado, clave pública SSH, huella de esa clave, fechas de creación y verificación, hash de las sesiones y hash del código de recuperación. Se usan para crear la cuenta, autenticarte, recuperar el acceso y prevenir abusos. No recibimos tu clave privada SSH ni almacenamos el código de recuperación en claro.
- **Mensajería:** contenido de mensajes, remitente, conversación, identificador de reintento, secuencia y fecha; membresías de canales; solicitudes y bloqueos de chats directos. Se usan para entregar y mostrar conversaciones, controlar el acceso y aplicar tus decisiones de aceptación o bloqueo.
- **Operación:** contadores agregados de peticiones y fallos y límites temporales de tasa en memoria. La aplicación no registra el cuerpo de mensajes, códigos ni tokens en sus logs operativos. El proveedor de infraestructura puede generar registros propios conforme a su servicio.

El correo no se muestra a otros usuarios. Los miembros de una conversación pueden leer los mensajes enviados a ella. El operador del servidor puede acceder técnicamente al contenido almacenado para operar o resolver incidentes. **Los mensajes no están cifrados de extremo a extremo.** HTTPS, WSS y SSH protegen su tránsito.

## Proveedores y transferencias

El servidor y PostgreSQL se alojarán en `[PROVEEDOR, REGIÓN Y PAÍS]`. Los respaldos cifrados se guardarán en `[PROVEEDOR, REGIÓN Y PAÍS]`. Los códigos de verificación y recuperación se enviarán mediante `[PROVEEDOR DE CORREO]`, que recibirá la dirección de destino y el texto necesario para entregar cada correo. Antes del lanzamiento se completará la información sobre ubicaciones y flujos internacionales de datos.

No vendemos datos personales ni usamos el contenido de mensajes para publicidad o entrenamiento de modelos. `[VERIFICAR CON LOS CONTRATOS Y CONFIGURACIONES REALES DE LOS PROVEEDORES]`.

## Conservación y borrado

- Los mensajes dejan de estar disponibles en el historial a las **720 horas (30 días)** de su creación. La purga física de la base activa se ejecuta al arrancar y cada hora en modo público; puede haber un breve desfase operativo tras el límite de acceso.
- Las solicitudes de chat directo vencen a los **7 días**. Los desafíos de correo vencen a los **10 minutos**. Las sesiones temporales vencen a las **8 horas**; sus registros caducados se limpiarán automáticamente antes de publicar esta política.
- El usuario puede borrar su cuenta desde el cliente. La operación borra su correo, claves, sesiones, membresías y mensajes que escribió de la base activa. Los mensajes escritos por otras personas se conservan hasta su propio vencimiento; los chats directos afectados quedan en solo lectura para ellas.
- Los respaldos completos se cifran antes de subirlos. La operación propuesta los elimina a los **3 días** y configura el mismo plazo en el proveedor. Un mensaje ya vencido o una cuenta borrada pueden permanecer en un respaldo por un período adicional, con un máximo público previsto de **7 días**. Este máximo debe verificarse contra el borrado real del proveedor antes de publicar esta política.
- El proveedor de correo puede conservar registros de entrega según sus condiciones. `[CONFIRMAR Y PRECISAR RETENCIÓN DEL PROVEEDOR]`.

## Tus opciones y derechos

Puedes rechazar solicitudes de chat directo, bloquear usuarios y borrar tu cuenta. Para ejercer derechos de acceso, rectificación, cancelación u oposición sobre tus datos, escribe al correo de privacidad indicado arriba. Podemos pedir una prueba razonable de que controlas la cuenta antes de responder. La [ANPD explica los derechos ARCO](https://www.gob.pe/9270-que-son-los-derechos-arco) y las vías de reclamo disponibles en Perú.

## Seguridad y cambios

Protegemos el acceso con claves SSH, correo verificado y un código de recuperación que se muestra una sola vez. Conserva tu clave privada y ese código fuera del alcance de terceros. Si pierdes tanto el acceso al correo como el código de recuperación, el flujo automático no puede recuperar tu cuenta. Publicaremos los cambios materiales de esta política en `[URL PÚBLICA]` antes de que entren en vigor.

## Revisión previa a publicación

La [ANPD indica que los titulares de bancos de datos personales, incluidas personas naturales, deben inscribirlos](https://www.gob.pe/institucion/anpd/pages/8060-inscribir-banco-de-datos-en-el-registro-nacional-de-proteccion-de-datos-personales). La [comunicación de flujos transfronterizos](https://www.gob.pe/9253-inscribir-flujo-transfronterizo-de-datos-personales) también debe revisarse al confirmar las ubicaciones de infraestructura y correo. Estas gestiones corresponden al titular y requieren sus datos personales; este borrador no afirma que ya se hayan completado.
