package cmd

import (
	"log"
	"net/http"
	"text/template"

	"activity-heatmap/internal/style"
	"activity-heatmap/internal/tiles"
)

func Execute() {
	mapStyle := style.GetMapStyle()

	if err := tiles.Generate(); err != nil {
		log.Fatalf("Failed to generate tiles: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tmpl, _ := template.ParseFiles("./web/index.html")
		tmpl.Execute(w, map[string]string{"MapStyle": mapStyle})
	})

	//go tiles.Watch()

	http.Handle("/favicon.svg", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/tiles/", tiles.ServePNGTile)
	http.HandleFunc("/api/heatmap", tiles.GetHeatmapGeoJSON)
	addr := ":3465"
	log.Println("Server started at port" + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}