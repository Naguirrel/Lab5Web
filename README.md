Lab 5 – Tracker de Series

Descripción
Aplicación web desarrollada en Go utilizando sockets TCP y SQLite. Permite llevar el control de series vistas, episodios actuales y progreso total.

El servidor implementa manejo manual de HTTP (sin frameworks), incluyendo parsing de requests, manejo de métodos GET/POST y generación dinámica de HTML.

Funcionalidades Base

Manejo correcto de métodos GET y POST

Parseo manual de application/x-www-form-urlencoded

Inserción en base de datos SQLite

Implementación del patrón POST / Redirect / GET (303 See Other)

Tabla generada dinámicamente desde la base de datos

Manejo de errores en consultas e inserciones

Uso de defer rows.Close() y defer db.Close()

Challenges Implementados

Estilos y CSS

Diseño con estilos personalizados

Hover en filas

Toast dinámico

Barra de progreso animada

Barra de progreso

Cálculo automático del porcentaje (episodios vistos / total)

Animación visual en cada fila

Texto especial para serie completa

Se muestra "✔ COMPLETA" cuando current_episode >= total_episodes

Botón +1

Incrementa episodio usando fetch() con método POST

Ruta /update

Validación para no exceder el total de episodios

Botón -1

Decrementa episodio usando fetch() con método POST

Ruta /decrement

Validación para no bajar de 1

Favicon

Implementado y servido manualmente desde el servidor

Ruta /favicon.ico

Ordenamiento por columna

Orden dinámico por ID

Orden por Nombre

Orden por Episodios actuales

Orden por Total de episodios

Orden por Progreso

Búsqueda en vivo (client-side)

Filtrado dinámico por nombre de serie

Rutas Implementadas

GET / → Tabla principal

GET /create → Formulario para nueva serie

POST /create → Inserta serie en SQLite y redirige

POST /update?id=X → Incrementa episodio

POST /decrement?id=X → Decrementa episodio

GET /favicon.ico → Sirve favicon

Tecnologías Utilizadas

Go (net, database/sql, bufio)

SQLite (modernc.org/sqlite)

JavaScript (fetch API)

HTML + CSS

Cómo Ejecutar
Ejecutar: go run main.go
Abrir en navegador: http://localhost:8080

ScreenShot:
<img width="1919" height="1078" alt="image" src="https://github.com/user-attachments/assets/5af503ea-1dd5-4d6b-b05e-aeb87092ba37" />


Notas
El proyecto implementa manejo manual del protocolo HTTP sin frameworks, con lectura explícita de headers y Content-Length, reforzando la comprensión del funcionamiento interno de un servidor web.
