Lab 5 – Tracker de Series

Descripción
Aplicación web desarrollada en Go utilizando sockets TCP y SQLite. Permite llevar el control de series vistas, episodios actuales, progreso total y sistema de calificación por estrellas.

El servidor implementa manejo manual de HTTP (sin frameworks), incluyendo parsing manual de requests, manejo de métodos GET/POST y generación dinámica de HTML.


Funcionalidades Base

* Manejo correcto de métodos GET y POST
* Parseo manual de application/x-www-form-urlencoded
* Inserción en base de datos SQLite
* Implementación del patrón POST / Redirect / GET (303 See Other)
* Tabla generada dinámicamente desde la base de datos
* Manejo de errores en consultas e inserciones
* Uso de defer rows.Close() y defer db.Close()


Challenges Implementados

Estilos y CSS — 10 pts
Diseño personalizado con estilos propios, hover en filas, barra visual y toast dinámico.

Barra de progreso — 15 pts
Cálculo automático del porcentaje (episodios vistos / total) y animación visual en cada fila.

Texto especial para serie completa — 10 pts
Se muestra "✔ COMPLETA" cuando current_episode >= total_episodes.

Botón -1 — 10 pts
Decrementa episodio usando fetch() con método POST en la ruta /decrement, validando que no baje de 1.

Favicon — 15 pts
Implementado y servido manualmente desde el servidor en la ruta /favicon.ico.

Ordenamiento por columna — 20 pts
Orden dinámico por ID, Nombre, Episodios actuales, Total de episodios y Progreso.

Sistema de rating con tabla propia — 40 pts
Implementación de tabla independiente "ratings" en SQLite.
Relación con tabla "series" mediante clave foránea.
Uso de INSERT con ON CONFLICT DO UPDATE.
Interfaz de 5 estrellas clickeables que guardan y actualizan el rating en base de datos.


Total de Challenges Completados

10 + 15 + 10 + 10 + 15 + 20 + 40 = 120 puntos acumulados


Rutas Implementadas

GET /
Tabla principal con listado dinámico de series.

GET /create
Formulario para crear nueva serie.

POST /create
Inserta nueva serie en SQLite y redirige con 303.

POST /update?id=X
Incrementa episodio actual con validación.

POST /decrement?id=X
Decrementa episodio actual con validación.

POST /rate?id=X
Guarda o actualiza el rating de la serie.

GET /favicon.ico
Sirve el favicon manualmente.


Tecnologías Utilizadas

Go (net, database/sql, bufio)
SQLite (modernc.org/sqlite)
JavaScript (fetch API)
HTML + CSS

Cómo Ejecutar
Ejecutar: go run main.go
Abrir en navegador: http://localhost:8080

ScreenShot:
<img width="1919" height="1079" alt="image" src="https://github.com/user-attachments/assets/f0a43660-f0a2-46b4-9595-dc3731141ccd" />


Notas
El proyecto implementa manejo manual del protocolo HTTP sin frameworks, con lectura explícita de headers y Content-Length, reforzando la comprensión del funcionamiento interno de un servidor web.
