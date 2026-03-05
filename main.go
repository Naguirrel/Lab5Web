package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"html"
	"log"
	"net"
	"strings"

	"io"
	"net/url"
	"os"
	"strconv"

	_ "modernc.org/sqlite"
)

type Serie struct {
	ID            int
	Name          string
	Current       int
	TotalEpisodes int
	Rating        int
}

func main() {
	db, err := sql.Open("sqlite", "file:series.db")
	if err != nil {
		log.Fatal("Error abriendo DB:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("No se pudo conectar a la DB (Ping):", err)
	}

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error escuchando en :8080:", err)
	}
	defer ln.Close()

	log.Println("Servidor escuchando en http://localhost:8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error aceptando conexión:", err)
			continue
		}
		go handleConn(conn, db)
	}
}

func handleConn(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Leer primera línea del request
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	line = strings.TrimRight(line, "\r\n")
	parts := strings.Fields(line)

	method := ""
	path := "/"

	if len(parts) >= 2 {
		method = parts[0]
		path = parts[1]
	}

	if path == "/favicon.ico" && method == "GET" {
		data, err := os.ReadFile("favicon.ico")
		if err != nil {
			return
		}

		resp := fmt.Sprintf(
			"HTTP/1.1 200 OK\r\n"+
				"Content-Type: image/x-icon\r\n"+
				"Content-Length: %d\r\n"+
				"Connection: close\r\n\r\n",
			len(data),
		)

		conn.Write([]byte(resp))
		conn.Write(data)
		return
	}

	// GET /create -> mostrar formulario
	if path == "/create" && method == "GET" {
		body := `<!doctype html>
  <html>
  <head>
    <meta charset="utf-8">
    <title>Crear Serie</title>
    <style>
      body { font-family: Arial, sans-serif; margin: 40px; }
      form { max-width: 400px; display: flex; flex-direction: column; gap: 12px; }
      input { padding: 8px; font-size: 14px; }
      button { padding: 10px; font-size: 14px; cursor: pointer; }
      a { display:inline-block; margin-top:15px; }
      label {
            font-weight: bold;
            margin-top: 5px;
            }
    </style>
  </head>
  <body>
    <h1>Agregar nueva serie</h1>

    <form method="POST" action="/create">
  
      <label for="series_name">Nombre de la serie</label>
      <input type="text" id="series_name" name="series_name" required>

      <label for="current_episode">Episodio actual</label>
      <input type="number" id="current_episode" name="current_episode" min="1" value="1" required>

      <label for="total_episodes">Total de episodios</label>
      <input type="number" id="total_episodes" name="total_episodes" min="1" required>

      <button type="submit">Guardar</button>
    </form>

    <a href="/">← Volver al listado</a>
  </body>
  </html>`

		writeHTTP(conn, "200 OK", body)
		return
	}

	if strings.HasPrefix(path, "/decrement") && method == "POST" {

		parts := strings.SplitN(path, "?", 2)

		if len(parts) > 1 {
			params, _ := url.ParseQuery(parts[1])
			id := params.Get("id")

			_, err := db.Exec(
				`UPDATE series
			 SET current_episode = current_episode - 1
			 WHERE id = ? AND current_episode > 1`,
				id,
			)

			if err != nil {
				writeHTTP(conn, "500 Internal Server Error", "error")
				return
			}

			resp := "HTTP/1.1 200 OK\r\n" +
				"Content-Type: text/plain\r\n" +
				"Content-Length: 2\r\n" +
				"Connection: close\r\n\r\nok"

			conn.Write([]byte(resp))
			return
		}
	}

	// POST /create -> recibir formulario
	if path == "/create" && method == "POST" {

		var contentLength int

		// Leer headers hasta línea vacía
		for {
			hLine, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			hLine = strings.TrimRight(hLine, "\r\n")

			// Línea vacía = fin de headers
			if hLine == "" {
				break
			}

			// Buscar Content-Length
			if strings.HasPrefix(hLine, "Content-Length:") {
				lengthStr := strings.TrimSpace(strings.TrimPrefix(hLine, "Content-Length:"))
				contentLength, _ = strconv.Atoi(lengthStr)
			}
		}

		// Leer exactamente Content-Length bytes
		bodyBytes := make([]byte, contentLength)
		_, err := io.ReadFull(reader, bodyBytes)
		if err != nil {
			return
		}

		body := string(bodyBytes)

		// Parsear application/x-www-form-urlencoded
		values, err := url.ParseQuery(body)
		if err != nil {
			log.Println("Error parseando form:", err)
		}

		name := values.Get("series_name")
		currentEp := values.Get("current_episode")
		totalEps := values.Get("total_episodes")

		// Convertir a enteros
		currentInt, err1 := strconv.Atoi(currentEp)
		totalInt, err2 := strconv.Atoi(totalEps)

		if err1 != nil || err2 != nil {
			log.Println("Error convirtiendo a int")
			writeHTTP(conn, "400 Bad Request", "<h1>Datos inválidos</h1>")
			return
		}

		// INSERT en la base
		_, err = db.Exec(
			"INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
			name, currentInt, totalInt,
		)

		if err != nil {
			log.Println("Error insertando en DB:", err)
			writeHTTP(conn, "500 Internal Server Error", "<h1>Error guardando en base</h1>")
			return
		}

		// Redirección 303 (POST/Redirect/GET)
		resp := "HTTP/1.1 303 See Other\r\n" +
			"Location: /\r\n" +
			"Content-Length: 0\r\n" +
			"Connection: close\r\n" +
			"\r\n"

		conn.Write([]byte(resp))
		return

		writeHTTP(conn, "200 OK", resp)
		return
	}

	// POST /update?id=3
	if strings.HasPrefix(path, "/update") && method == "POST" {

		parts := strings.SplitN(path, "?", 2)
		route := parts[0]

		if route == "/update" && len(parts) > 1 {

			params, _ := url.ParseQuery(parts[1])
			id := params.Get("id")

			_, err := db.Exec(
				`UPDATE series
        SET current_episode = current_episode + 1
        WHERE id = ? AND current_episode < total_episodes`,
				id,
			)

			if err != nil {
				writeHTTP(conn, "500 Internal Server Error", "error")
				return
			}

			resp := "HTTP/1.1 200 OK\r\n" +
				"Content-Type: text/plain\r\n" +
				"Content-Length: 2\r\n" +
				"Connection: close\r\n" +
				"\r\n" +
				"ok"

			conn.Write([]byte(resp))
			return
		}
	}

	if strings.HasPrefix(path, "/rate") && method == "POST" {

		parts := strings.SplitN(path, "?", 2)
		if len(parts) > 1 {

			params, _ := url.ParseQuery(parts[1])
			id := params.Get("id")

			var contentLength int

			for {
				hLine, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				hLine = strings.TrimRight(hLine, "\r\n")
				if hLine == "" {
					break
				}
				if strings.HasPrefix(hLine, "Content-Length:") {
					lengthStr := strings.TrimSpace(strings.TrimPrefix(hLine, "Content-Length:"))
					contentLength, _ = strconv.Atoi(lengthStr)
				}
			}

			bodyBytes := make([]byte, contentLength)
			_, err := io.ReadFull(reader, bodyBytes)
			if err != nil {
				return
			}

			values, _ := url.ParseQuery(string(bodyBytes))
			score := values.Get("score")

			_, err = db.Exec(`
        INSERT INTO ratings (serie_id, score)
        VALUES (?, ?)
        ON CONFLICT(serie_id)
        DO UPDATE SET score = excluded.score;
      `, id, score)

			if err != nil {
				writeHTTP(conn, "500 Internal Server Error", "error")
				return
			}

			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length:2\r\n\r\nok"))
			return
		}
	}

	// Solo atendemos "/" con la tabla. Cualquier otro path, mostramos un mensaje simple.
	if path == "/" {
		htmlBody, status := renderSeriesPage(db)
		writeHTTP(conn, status, htmlBody)
		return
	}

	// Página simple para otros paths (cumple con "mostrar el path")
	body := fmt.Sprintf(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>mini server</title></head>
<body>
  <h1>Hola</h1>
  <p>You requested: <b>%s</b></p>
  <p>Tip: visita <a href="/">/</a> para ver tu tabla de series.</p>
</body>
</html>`, html.EscapeString(path))

	writeHTTP(conn, "200 OK", body)
}

func renderSeriesPage(db *sql.DB) (string, string) {
	rows, err := db.Query(`SELECT s.id, s.name, s.current_episode, s.total_episodes,
                        COALESCE(r.score, 0)
                        FROM series s
                        LEFT JOIN ratings r ON s.id = r.serie_id
                        ORDER BY s.id`)
	if err != nil {
		body := fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"><title>Error</title></head>
<body><h1>Error consultando la base de datos</h1><pre>%s</pre></body></html>`,
			html.EscapeString(err.Error()))
		return body, "500 Internal Server Error"
	}
	defer rows.Close()

	var series []Serie
	for rows.Next() {
		var s Serie
		if err := rows.Scan(&s.ID, &s.Name, &s.Current, &s.TotalEpisodes, &s.Rating); err != nil {
			body := fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"><title>Error</title></head>
<body><h1>Error leyendo filas</h1><pre>%s</pre></body></html>`,
				html.EscapeString(err.Error()))
			return body, "500 Internal Server Error"
		}
		series = append(series, s)
	}
	if err := rows.Err(); err != nil {
		body := fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"><title>Error</title></head>
<body><h1>Error final</h1><pre>%s</pre></body></html>`,
			html.EscapeString(err.Error()))
		return body, "500 Internal Server Error"
	}

	// filas con data-current/total para JS
	var tableRows strings.Builder
	for _, s := range series {
		name := html.EscapeString(s.Name)
		statusText := ""
		if s.Current >= s.TotalEpisodes {
			statusText = `<span style="color:green;font-weight:bold;">✔ COMPLETA</span>`
		}
		color1 := "#ccc"
		if s.Rating >= 1 {
			color1 = "gold"
		}

		color2 := "#ccc"
		if s.Rating >= 2 {
			color2 = "gold"
		}

		color3 := "#ccc"
		if s.Rating >= 3 {
			color3 = "gold"
		}

		color4 := "#ccc"
		if s.Rating >= 4 {
			color4 = "gold"
		}

		color5 := "#ccc"
		if s.Rating >= 5 {
			color5 = "gold"
		}

		tableRows.WriteString(fmt.Sprintf(
			`<tr class="row" data-name="%s" data-current="%d" data-total="%d">
          <td>%d</td>
          <td class="name">%s %s</td>
          <td>%d</td>
          <td>%d</td>
          <td class="progressCell"></td>

          <td>
            <span onclick="rate(%d,1)" style="cursor:pointer;color:%s;">★</span>
            <span onclick="rate(%d,2)" style="cursor:pointer;color:%s;">★</span>
            <span onclick="rate(%d,3)" style="cursor:pointer;color:%s;">★</span>
            <span onclick="rate(%d,4)" style="cursor:pointer;color:%s;">★</span>
            <span onclick="rate(%d,5)" style="cursor:pointer;color:%s;">★</span>
          </td>

          <td>
            <button onclick="nextEpisode(%d)">+1</button>
            <button onclick="prevEpisode(%d)">-1</button>
          </td>
        </tr>`,

			// ====== PARÁMETROS EXACTOS EN ORDEN ======

			strings.ToLower(name), // %s
			s.Current,             // %d
			s.TotalEpisodes,       // %d

			s.ID,       // %d
			name,       // %s
			statusText, // %s

			s.Current,       // %d
			s.TotalEpisodes, // %d

			s.ID, color1, // rate 1
			s.ID, color2, // rate 2
			s.ID, color3, // rate 3
			s.ID, color4, // rate 4
			s.ID, color5, // rate 5

			s.ID, // nextEpisode
			s.ID, // prevEpisode
		))
	}

	body := fmt.Sprintf(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Tracker</title>
  <link rel="icon" href="/favicon.ico">
  <style>
    body { font-family: Arial, sans-serif; margin: 40px; }
    h1 { margin: 0 0 8px 0; }
    .topbar { display:flex; gap:12px; align-items:center; margin: 12px 0 18px; flex-wrap:wrap; }
    input[type="text"] { padding:10px 12px; border:1px solid #ccc; border-radius:10px; min-width:260px; }
    .pill { padding:8px 12px; border:1px solid #ddd; border-radius:999px; background:#fafafa; font-size: 13px; }
    table { border-collapse: collapse; width: 100%%; max-width: 980px; }
    th, td { border: 1px solid #ddd; padding: 10px; text-align: left; }
    th { background: #f5f5f5; cursor:pointer; user-select:none; position:relative; }
    th .arrow { font-size: 12px; margin-left: 6px; opacity: .6; }
    tr.row:hover { background: #fffbe6; transition: background .15s ease; }
    .progressWrap { width: 220px; }
    .barOuter { height: 10px; background:#eee; border-radius:999px; overflow:hidden; }
    .barInner { height: 10px; width:0%%; border-radius:999px; background: #4caf50; transition: width .6s ease; }
    .pct { font-size: 12px; color:#555; margin-top:6px; }
    .hidden { display:none; }
    .toast {
      position: fixed; bottom: 18px; right: 18px;
      background: #111; color: #fff; padding: 10px 12px;
      border-radius: 12px; opacity: 0; transform: translateY(8px);
      transition: all .2s ease; font-size: 13px;
    }
    .toast.show { opacity: 0.95; transform: translateY(0); }
  </style>
</head>
<body>
  <h1>Tracker de Series</h1>
  <p><a href="/create">+ Agregar nueva serie</a></p>

  <div class="topbar">
    <input id="search" type="text" placeholder="Buscar serie... (filtra en vivo)">
    <span class="pill" id="countPill">Mostrando 0</span>
  </div>

  <table id="tbl">
    <thead>
      <tr>
        <th data-col="0"># <span class="arrow" id="a0"></span></th>
        <th data-col="1">Name <span class="arrow" id="a1"></span></th>
        <th data-col="2">Current <span class="arrow" id="a2"></span></th>
        <th data-col="3">Total <span class="arrow" id="a3"></span></th>
        <th data-col="4">Progress <span class="arrow" id="a4"></span></th>
        <th>Rating</th>
        <th>Episodios vistos</th>
      </tr>
    </thead>
    <tbody>
      %s
    </tbody>
  </table>

  <div class="toast" id="toast"></div>

<script>
(function () {
  const tbl = document.getElementById("tbl");
  const tbody = tbl.querySelector("tbody");
  const search = document.getElementById("search");
  const countPill = document.getElementById("countPill");
  const toast = document.getElementById("toast");

  function showToast(msg) {
    toast.textContent = msg;
    toast.classList.add("show");
    clearTimeout(showToast._t);
    showToast._t = setTimeout(function () {
      toast.classList.remove("show");
    }, 1200);
  }

  // 1) Crear barras de progreso animadas
  const rows = Array.from(tbody.querySelectorAll("tr.row"));
  rows.forEach(function (r) {
    const cur = parseInt(r.dataset.current, 10);
    const tot = parseInt(r.dataset.total, 10);
    const pct = tot > 0 ? Math.min(100, Math.round((cur / tot) * 100)) : 0;

    const cell = r.querySelector(".progressCell");
    cell.innerHTML =
      '<div class="progressWrap">' +
        '<div class="barOuter"><div class="barInner"></div></div>' +
        '<div class="pct">' + pct + '%%</div>' +
      '</div>';

    setTimeout(function () {
      const bar = cell.querySelector(".barInner");
      if (bar) bar.style.width = pct + "%%";
    }, 60);

    // click en fila: toast
    r.addEventListener("click", function () {
      const nm = r.querySelector(".name") ? r.querySelector(".name").textContent : "Serie";
      showToast(nm + " - " + pct + "%% completado");
    });
  });

  // 2) Buscador en vivo
  function applyFilter() {
    const q = search.value.trim().toLowerCase();
    let visible = 0;
    rows.forEach(function (r) {
      const name = r.dataset.name || "";
      const match = name.indexOf(q) !== -1;
      r.classList.toggle("hidden", !match);
      if (match) visible++;
    });
    countPill.textContent = "Mostrando " + visible + " / " + rows.length;
  }
  search.addEventListener("input", applyFilter);
  applyFilter();

  // 3) Ordenar al click en encabezados
  let sortState = { col: 0, asc: true };

  function getCellValue(row, col) {
    const tds = row.querySelectorAll("td");
    if (!tds[col]) return "";
    const text = tds[col].textContent.trim();

    if (col === 0 || col === 2 || col === 3) return parseFloat(text);
    if (col === 4) {
      const cur = parseInt(row.dataset.current, 10);
      const tot = parseInt(row.dataset.total, 10);
      return tot > 0 ? (cur / tot) : 0;
    }
    return text.toLowerCase();
  }

  function setArrows() {
    for (let i = 0; i <= 4; i++) {
      const a = document.getElementById("a" + i);
      if (a) a.textContent = "";
    }
    const a = document.getElementById("a" + sortState.col);
    if (a) a.textContent = sortState.asc ? "▲" : "▼";
  }

  tbl.querySelectorAll("th").forEach(function (th) {
    th.addEventListener("click", function () {
      const col = parseInt(th.dataset.col, 10);
      if (sortState.col === col) sortState.asc = !sortState.asc;
      else sortState = { col: col, asc: true };

      const sorted = rows.slice().sort(function (ra, rb) {
        const A = getCellValue(ra, col);
        const B = getCellValue(rb, col);

        const aNum = typeof A === "number" && !isNaN(A);
        const bNum = typeof B === "number" && !isNaN(B);

        if (aNum && bNum) return sortState.asc ? (A - B) : (B - A);
        if (A < B) return sortState.asc ? -1 : 1;
        if (A > B) return sortState.asc ? 1 : -1;
        return 0;
      });

      sorted.forEach(function (r) {
        tbody.appendChild(r);
      });

      setArrows();
      showToast("Ordenado por " + th.textContent.trim());
      applyFilter();
    });
  });

  setArrows();
})();

async function nextEpisode(id) {
    const url = "/update?id=" + id;

    const response = await fetch(url, { method: "POST" });

    if (response.ok) {
        location.reload();
    }
}

async function prevEpisode(id) {
    const url = "/decrement?id=" + id;

    const response = await fetch(url, { method: "POST" });

    if (response.ok) {
        location.reload();
    }
}

async function rate(id, score) {
    const response = await fetch("/rate?id=" + id, {
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        },
        body: "score=" + score
    });

    if (response.ok) {
        location.reload();
    }
}
</script>

</body>
</html>`, tableRows.String())

	return body, "200 OK"
}

func writeHTTP(conn net.Conn, status string, body string) {
	// Content-Length en bytes
	contentLength := len([]byte(body))

	resp := fmt.Sprintf(
		"HTTP/1.1 %s\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"Content-Length: %d\r\n"+
			"Connection: close\r\n"+
			"\r\n"+
			"%s",
		status, contentLength, body,
	)

	_, _ = conn.Write([]byte(resp))
}
