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

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type result struct {
	latency time.Duration
	err     error
}

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

	msg := Message{Action: "ping", Data: json.RawMessage("null")}
	var resp Message

	for i := 0; i < requestsPerConn; i++ {
		if timeout > 0 {
			_ = conn.SetDeadline(time.Now().Add(timeout))
		}

		start := time.Now()
		if err := enc.Encode(&msg); err != nil {
			out <- result{err: fmt.Errorf("worker %d encode: %w", id, err)}
			return
		}
		if err := dec.Decode(&resp); err != nil {
			out <- result{err: fmt.Errorf("worker %d decode: %w", id, err)}
			return
		}
		if resp.Action != "pong_response" {
			out <- result{err: fmt.Errorf("worker %d resposta inesperada: %s", id, resp.Action)}
			return
		}
		out <- result{latency: time.Since(start)}
	}
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
	addr := getenv("SERVER_ADDR", "witchcraft-server:8080")
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
}
