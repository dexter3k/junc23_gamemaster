package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "strconv"
    "log"
    "math/rand"

    "github.com/gorilla/mux"
    "github.com/gorilla/handlers"
)

// Game represents a match session
type Game struct {
    Code     string
    Players  map[string]int
    JoinChan chan int
    Scores   map[string]int
    ScoreChan chan bool

    mut sync.Mutex
}

var (
    games      = make(map[string]*Game)
    gamesMutex = &sync.Mutex{}
)

func createGame(w http.ResponseWriter, r *http.Request) {
    gamesMutex.Lock()
    defer gamesMutex.Unlock()

    var gameCode string
    for {
        gameCode = fmt.Sprintf("%06d", rand.Intn(1000000)) // generates a number between 000000 and 999999
        if _, exists := games[gameCode]; !exists {
            break
        }
    }

    game := &Game{
        Code:     gameCode,
        Players:  make(map[string]int),
        JoinChan: make(chan int, 2), // buffered channel
        Scores:   make(map[string]int),
        ScoreChan: make(chan bool, 2), // buffered channel
    }
    games[gameCode] = game

    log.Printf("createGame(void) -> %q\n", gameCode)

    json.NewEncoder(w).Encode(map[string]string{"gameCode": gameCode})
}

func joinGame(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    gameCode := vars["gameId"]

    gamesMutex.Lock()
    game, exists := games[gameCode]
    gamesMutex.Unlock()

    if !exists {
        http.Error(w, "Game not found", http.StatusNotFound)
        return
    }

    game.mut.Lock()
    userID := len(game.Players) + 1
    game.Players[fmt.Sprintf("user%d", userID)] = 0
    game.mut.Unlock()
    if userID == 2 {
        close(game.JoinChan) // Close channel to notify that both players have joined
    }

    <-game.JoinChan // Wait until both players have joined

    json.NewEncoder(w).Encode(map[string]int{"userId": userID})
}

func completeGame(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    gameCode := vars["gameId"]
    userID := r.URL.Query().Get("user_id")
    score, _ := strconv.Atoi(r.URL.Query().Get("score"))

    gamesMutex.Lock()
    game, exists := games[gameCode]
    gamesMutex.Unlock()

    if !exists {
        http.Error(w, "Game not found", http.StatusNotFound)
        return
    }

    game.mut.Lock()
    game.Scores[userID] = score
    if len(game.Scores) == 2 {
        close(game.ScoreChan) // Close channel to notify that both scores are submitted
    }
    game.mut.Unlock()

    <-game.ScoreChan // Wait until both scores are submitted

    winner := determineWinner(game.Scores)
    won := winner == userID

    json.NewEncoder(w).Encode(map[string]bool{"won": won})
}

func determineWinner(scores map[string]int) string {
    maxScore := -1
    winner := ""
    for user, score := range scores {
        if score > maxScore {
            maxScore = score
            winner = user
        }
    }
    return winner
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/createGame", createGame).Methods("POST")
    r.HandleFunc("/joinGame/{gameId}", joinGame).Methods("POST")
    r.HandleFunc("/completeGame/{gameId}", completeGame).Methods("POST")

    headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
    originsOk := handlers.AllowedOrigins([]string{"*"})
    methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

    http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(r))
}
