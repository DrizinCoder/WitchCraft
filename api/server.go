package api

import (
	"WitchCraft/Cards"
	"WitchCraft/Player"
	"encoding/json"
	"net/http"
)

var playerManager = Player.NewManager()
var stock = Cards.NewStock()

func Setup() {

	stock.CreateCard("Fireball", 10, 5, Cards.GOLD)
	stock.CreateCard("Icebolt", 8, 6, Cards.SILVER)
	stock.CreateCard("Goblin", 5, 10, Cards.BRONZE)
	stock.CreateCard("Dragon", 20, 20, Cards.DIAMOND)
	stock.CreateCard("Knight", 12, 15, Cards.SILVER)
	stock.CreateCard("Elf", 7, 8, Cards.BRONZE)

	http.HandleFunc("/player/create", createPlayerHandler)
	http.HandleFunc("/player/login", loginPlayerHlander)
	http.HandleFunc("/player/openpack", openPackHandler)
	http.HandleFunc("/player/getplayer", getPlayerHandler)

	http.ListenAndServe(":8080", nil)
}

func createPlayerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserName string `json:"username"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	player := playerManager.Create_Player(req.UserName, req.Login, req.Password)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func loginPlayerHlander(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	player, err := playerManager.Login(req.Login, req.Password)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(player)
}

func openPackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PlayerId int `json:"id"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	pack, err := playerManager.Open_pack(req.PlayerId, stock)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}

	json.NewEncoder(w).Encode(pack)

}

func getPlayerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PlayerID int `json:"id"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	player, err := playerManager.Search_Player(req.PlayerID)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}

	json.NewEncoder(w).Encode(player)
}

/*
---oq falta implementar---

-Enqueue
-Create_Card
*/

/*
curl -X POST -d '{"username":"Gui","login":"gui123","password":"123"}' http://localhost:8080/player/create

curl -X POST -d '{"login":"gui123","password":"123"}' http://localhost:8080/player/login

curl -X POST -d '{"id":1}' http://localhost:8080/player/openpack

curl -X GET -d '{"id":1}' http://localhost:8080/player/getplayer
*/
