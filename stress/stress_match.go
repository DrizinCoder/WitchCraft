package stress

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

// ===== Controle de duplicatas de match =====
var (
	globalMatches = make(map[int]int) // playerID -> matchID
)

// ==========================================

func getenvMatch(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func atoiEnvMatch(key, def string) int {
	v := getenvMatch(key, def)
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

func durationMsEnvMatch(key, def string) time.Duration {
	n := atoiEnvMatch(key, def)
	if n <= 0 {
		return 0
	}
	return time.Duration(n) * time.Millisecond
}

func percentileMatch(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	rank := int(p*float64(len(sorted)) + 0.5)
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}

// Worker de matchmaking (modelo igual ao stress de cartas)
func workerMatches(id int, addr string, matchesPerWorker int, timeout time.Duration, out chan<- result, rampDelay time.Duration) {
	if rampDelay > 0 && id > 0 {
		time.Sleep(rampDelay * time.Duration(id))
	}

	dialTimeout := timeout
	if dialTimeout == 0 {
		dialTimeout = 3 * time.Second
	}

	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		out <- result{err: fmt.Errorf("worker %d dial: %w", id, err)}
		return
	}
	defer conn.Close()

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	username := fmt.Sprintf("match_user_%d_%d", id, time.Now().UnixNano())
	password := "pwd"

	// 1) create_player
	createPayload := map[string]string{
		"username": username,
		"login":    username,
		"password": password,
	}
	req := Message{Action: "create_player", Data: toRaw(createPayload)}
	if err := enc.Encode(&req); err != nil {
		out <- result{err: fmt.Errorf("worker %d encode create_player: %w", id, err)}
		return
	}
	var resp Message
	if err := dec.Decode(&resp); err != nil {
		out <- result{err: fmt.Errorf("worker %d decode create_player resp: %w", id, err)}
		return
	}
	if resp.Action != "create_player_response" {
		out <- result{err: fmt.Errorf("worker %d unexpected create_player response: %s", id, resp.Action)}
		return
	}

	// 2) login_player
	loginPayload := map[string]string{
		"login":    username,
		"password": password,
	}
	req = Message{Action: "login_player", Data: toRaw(loginPayload)}
	if err := enc.Encode(&req); err != nil {
		out <- result{err: fmt.Errorf("worker %d encode login: %w", id, err)}
		return
	}
	if err := dec.Decode(&resp); err != nil {
		out <- result{err: fmt.Errorf("worker %d decode login resp: %w", id, err)}
		return
	}
	if resp.Action != "login_player_response" {
		out <- result{err: fmt.Errorf("worker %d unexpected login response: %s", id, resp.Action)}
		return
	}

	var loginResp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp.Data, &loginResp); err != nil {
		out <- result{err: fmt.Errorf("worker %d unmarshal login: %w", id, err)}
		return
	}
	playerID := loginResp.ID

	// 3) Enqueue várias vezes e aguardar match
	matchesCreated := 0
	for i := 0; i < matchesPerWorker; i++ {
		if timeout > 0 {
			_ = conn.SetDeadline(time.Now().Add(timeout))
		}

		start := time.Now()

		enqueueReq := Message{
			Action: "enqueue_player",
			Data:   toRaw(map[string]int{"id": playerID}),
		}
		if err := enc.Encode(&enqueueReq); err != nil {
			out <- result{err: fmt.Errorf("worker %d encode enqueue_player: %w", id, err)}
			return
		}

		// Loop para ler mensagens do servidor até receber um match
		for {
			var serverMsg Message
			if err := dec.Decode(&serverMsg); err != nil {
				return
			}

			switch serverMsg.Action {
			case "enqueue_response":
				// Só confirma que o jogador foi enfileirado
				continue
			case "Game_start":
				var matchResp struct {
					MatchID int   `json:"match_id"`
					Players []int `json:"players"`
				}
				if err := json.Unmarshal(serverMsg.Data, &matchResp); err != nil {
					out <- result{err: fmt.Errorf("worker %d unmarshal match_found: %w", id, err)}
					return
				}

				// Conta match
				matchesCreated++
				out <- result{latency: time.Since(start)}
				conn.Close()
			case "error_response":
				out <- result{err: fmt.Errorf("worker %d server error: %s", id, string(serverMsg.Data))}
				return
			default:
				// ignora mensagens inesperadas
			}
			if matchesCreated >= matchesPerWorker {
				break
			}
		}
	}
}

// Função de execução do stress de matchmaking
func RunMatch() {
	addr := getenvMatch("SERVER_ADDR", "127.0.0.1:8080")
	concurrency := atoiEnvMatch("STRESS_CONCURRENCY", "100")
	matchesPerWorker := atoiEnvMatch("STRESS_MATCHES", "1")
	timeout := durationMsEnvMatch("STRESS_TIMEOUT_MS", "2000")
	ramp := durationMsEnvMatch("STRESS_RAMP_MS", "0")

	totalExpected := concurrency * matchesPerWorker
	results := make(chan result, totalExpected)

	fmt.Printf("=== Matchmaking Stress Test ===\n")
	fmt.Printf("Alvo: %s\n", addr)
	fmt.Printf("Conexões: %d | Matches/Worker: %d | Timeout: %v | Ramp/worker: %v\n",
		concurrency, matchesPerWorker, timeout, ramp)

	startAll := time.Now()

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			workerMatches(id, addr, matchesPerWorker, timeout, results, ramp)
		}(i)
	}

	wg.Wait()
	close(results)

	elapsed := time.Since(startAll)

	latencies := make([]time.Duration, 0, totalExpected)
	errors := 0
	for r := range results {
		if r.err != nil {
			errors++
			fmt.Println("ERR:", r.err)
			continue
		}
		latencies = append(latencies, r.latency)
	}

	success := len(latencies)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	var avg time.Duration
	for _, d := range latencies {
		avg += d
	}
	if success > 0 {
		avg /= time.Duration(success)
	}

	p50 := percentileMatch(latencies, 0.50)
	p90 := percentileMatch(latencies, 0.90)
	p99 := percentileMatch(latencies, 0.99)
	qps := float64(success) / elapsed.Seconds()

	fmt.Printf("\n=== Resultado ===\n")
	fmt.Printf("Sucesso: %d | Erros: %d | Total: %d\n", success, errors, totalExpected)
	fmt.Printf("Duração total: %v\n", elapsed)
	if success > 0 {
		fmt.Printf("Média: %v | p50: %v | p90: %v | p99: %v\n", avg, p50, p90, p99)
		fmt.Printf("Throughput: %.2f req/s (QPS)\n", qps)
	}

	fmt.Printf("\n=== Análise de Match ===\n")
	fmt.Printf("📦 Total de jogadores pareados: %d\n", len(globalMatches))
}
