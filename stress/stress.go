package stress

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type result struct {
	latency time.Duration
	err     error
}

// ===== Controle de duplicatas globais =====
var (
	globalSeenCards = make(map[int]string) // ID -> Nome
	globalCardsMu   sync.Mutex
	duplicateCount  int64
	dupStats        = make(map[int]int) // ID -> vezes que apareceu
)

func checkDuplicate(workerID int, cardID int, cardName string) {
	globalCardsMu.Lock()
	defer globalCardsMu.Unlock()

	if _, exists := globalSeenCards[cardID]; exists {
		fmt.Printf("⚠️  Duplicata detectada pelo worker %d: %d (%s)\n", workerID, cardID, cardName)
		atomic.AddInt64(&duplicateCount, 1)
		dupStats[cardID]++
	} else {
		globalSeenCards[cardID] = cardName
		dupStats[cardID] = 1
	}
}

// ==========================================

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func atoiEnv(key, def string) int {
	v := getenv(key, def)
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

func durationMsEnv(key, def string) time.Duration {
	n := atoiEnv(key, def)
	if n <= 0 {
		return 0
	}
	return time.Duration(n) * time.Millisecond
}

func worker(id int, addr string, requestsPerConn int, timeout time.Duration, out chan<- result, rampDelay time.Duration) {
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

	// cria usuário único por worker
	unique := fmt.Sprintf("stress_user_%d_%d", id, time.Now().UnixNano())
	password := "pwd"

	// 1) create_player
	createPayload := map[string]string{
		"username": unique,
		"login":    unique,
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
		"login":    unique,
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

	// extrai player id do response
	var loginResp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp.Data, &loginResp); err != nil {
		out <- result{err: fmt.Errorf("worker %d unmarshal login response: %w", id, err)}
		return
	}
	playerID := loginResp.ID

	// 3) repetir open_pack requestsPerConn vezes
	openReq := Message{
		Action: "open_pack",
		Data:   toRaw(map[string]int{"id": playerID}),
	}

	for i := 0; i < requestsPerConn; i++ {
		if timeout > 0 {
			_ = conn.SetDeadline(time.Now().Add(timeout))
		}

		start := time.Now()
		if err := enc.Encode(&openReq); err != nil {
			out <- result{err: fmt.Errorf("worker %d encode open_pack: %w", id, err)}
			return
		}
		if err := dec.Decode(&resp); err != nil {
			out <- result{err: fmt.Errorf("worker %d decode open_pack: %w", id, err)}
			return
		}
		if resp.Action != "open_pack_response" {
			out <- result{err: fmt.Errorf("worker %d unexpected open_pack response: %s", id, resp.Action)}
			return
		}

		var cards []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(resp.Data, &cards); err != nil {
			out <- result{err: fmt.Errorf("worker %d unmarshal open_pack response: %w", id, err)}
			return
		}

		// checar duplicatas globais
		for _, c := range cards {
			checkDuplicate(id, c.ID, c.Name)
		}
		out <- result{latency: time.Since(start)}
	}
}

// helper
func toRaw(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}

func percentile(sorted []time.Duration, p float64) time.Duration {
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

func Run() {
	addr := getenv("SERVER_ADDR", "172.16.201.6:8080")
	concurrency := atoiEnv("STRESS_CONCURRENCY", "1000")
	requestsPerConn := atoiEnv("STRESS_REQUESTS", "100")
	timeout := durationMsEnv("STRESS_TIMEOUT_MS", "2000")
	ramp := durationMsEnv("STRESS_RAMP_MS", "0")

	totalExpected := concurrency * requestsPerConn
	results := make(chan result, totalExpected)

	fmt.Printf("=== WitchCraft Stress Test ===\n")
	fmt.Printf("Alvo: %s\n", addr)
	fmt.Printf("Conexões: %d | Reqs/Conexão: %d | Timeout: %v | Ramp/worker: %v\n",
		concurrency, requestsPerConn, timeout, ramp)

	startAll := time.Now()

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			worker(id, addr, requestsPerConn, timeout, results, ramp)
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

	p50 := percentile(latencies, 0.50)
	p90 := percentile(latencies, 0.90)
	p99 := percentile(latencies, 0.99)
	qps := float64(success) / elapsed.Seconds()

	fmt.Printf("\n=== Resultado ===\n")
	fmt.Printf("Sucesso: %d | Erros: %d | Total: %d\n", success, errors, totalExpected)
	fmt.Printf("Duração total: %v\n", elapsed)
	if success > 0 {
		fmt.Printf("Média: %v | p50: %v | p90: %v | p99: %v\n", avg, p50, p90, p99)
		fmt.Printf("Throughput: %.2f req/s (QPS)\n", qps)
	}

	// ===== LOG de duplicatas =====
	fmt.Printf("\n=== Análise de Cartas ===\n")
	fmt.Printf("🔁 Cartas duplicadas detectadas: %d\n", atomic.LoadInt64(&duplicateCount))
	fmt.Printf("📦 Total de cartas únicas vistas: %d\n", len(globalSeenCards))
}
