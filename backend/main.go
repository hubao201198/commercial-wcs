package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type ASN struct {
	ID        string `json:"id"`
	Supplier  string `json:"supplier"`
	SKU       string `json:"sku"`
	Expected  int    `json:"expected"`
	Received  int    `json:"received"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type Inventory struct {
	SKU       string `json:"sku"`
	Location  string `json:"location"`
	Available int    `json:"available"`
	Receiving int    `json:"receiving"`
	Allocated int    `json:"allocated"`
	Picked    int    `json:"picked"`
}

type Order struct {
	ID        string `json:"id"`
	Customer  string `json:"customer"`
	SKU       string `json:"sku"`
	Qty       int    `json:"qty"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type Task struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Ref      string `json:"ref"`
	Status   string `json:"status"`
	Executor string `json:"executor"`
}

type Audit struct {
	At     string `json:"at"`
	Action string `json:"action"`
	Detail string `json:"detail"`
	Actor  string `json:"actor"`
}

type State struct {
	sync.Mutex
	Mode   string
	Kill   bool
	ASNs   []ASN
	Inv    map[string]*Inventory
	Orders []Order
	Tasks  []Task
	Audits []Audit
	Seq    int
}

var s = &State{
	Mode: "TRADITIONAL_WMS",
	Inv: map[string]*Inventory{
		"COKE330":  {SKU: "COKE330", Location: "A01-01-01", Available: 120},
		"WATER550": {SKU: "WATER550", Location: "A01-02-01", Available: 80},
		"MILK250":  {SKU: "MILK250", Location: "B02-01-03", Available: 56},
	},
}

func main() {
	s.audit("SYSTEM_BOOT", "AI Warehouse OS demo initialized", "system")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		jsonOut(w, map[string]any{"ok": true, "service": "ai-warehouse-os", "time": time.Now().Format(time.RFC3339)})
	})
	mux.HandleFunc("/api/dashboard", dashboard)
	mux.HandleFunc("/api/mode", mode)
	mux.HandleFunc("/api/asns", asns)
	mux.HandleFunc("/api/asns/", asnAction)
	mux.HandleFunc("/api/inventory", inventory)
	mux.HandleFunc("/api/orders", orders)
	mux.HandleFunc("/api/orders/", orderAction)
	mux.HandleFunc("/api/tasks", tasks)
	mux.HandleFunc("/api/tasks/", taskAction)
	mux.HandleFunc("/api/audit", audit)
	mux.HandleFunc("/api/ai/plan", aiPlan)
	mux.Handle("/", spaHandler("./public"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("AI Warehouse OS listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, cors(mux)))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-Actor")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func spaHandler(root string) http.Handler {
	fs := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/health" {
			http.NotFound(w, r)
			return
		}
		path := root + r.URL.Path
		if r.URL.Path != "/" {
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
		}
		http.ServeFile(w, r, root+"/index.html")
	})
}

func actor(r *http.Request) string {
	if a := r.Header.Get("X-Actor"); a != "" {
		return a
	}
	return "demo.supervisor"
}

func jsonOut(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func bad(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	jsonOut(w, map[string]any{"error": msg})
}

func now() string { return time.Now().Format("2006-01-02 15:04:05") }

func (st *State) audit(action, detail, act string) {
	st.Audits = append([]Audit{{At: now(), Action: action, Detail: detail, Actor: act}}, st.Audits...)
	if len(st.Audits) > 200 {
		st.Audits = st.Audits[:200]
	}
}

func nextID(prefix string) string {
	s.Seq++
	return fmt.Sprintf("%s%05d", prefix, s.Seq)
}

func effectiveMode() string {
	if s.Kill {
		return "TRADITIONAL_WMS"
	}
	return s.Mode
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	var available, receiving, allocated, picked int
	for _, v := range s.Inv {
		available += v.Available
		receiving += v.Receiving
		allocated += v.Allocated
		picked += v.Picked
	}
	pending := 0
	for _, t := range s.Tasks {
		if t.Status != "SUCCEEDED" {
			pending++
		}
	}

	jsonOut(w, map[string]any{
		"mode":           effectiveMode(),
		"configuredMode": s.Mode,
		"killSwitch":     s.Kill,
		"available":      available,
		"receiving":      receiving,
		"allocated":      allocated,
		"picked":         picked,
		"orders":         len(s.Orders),
		"pendingTasks":   pending,
		"asns":           len(s.ASNs),
	})
}

func mode(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	if r.Method == http.MethodGet {
		jsonOut(w, map[string]any{"configuredMode": s.Mode, "effectiveMode": effectiveMode(), "killSwitch": s.Kill})
		return
	}
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Mode       string `json:"mode"`
		KillSwitch *bool  `json:"killSwitch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bad(w, "invalid request")
		return
	}
	if req.Mode != "" {
		if req.Mode != "TRADITIONAL_WMS" && req.Mode != "AI_WAREHOUSE_OS" {
			bad(w, "invalid mode")
			return
		}
		s.Mode = req.Mode
		s.audit("MODE_CHANGED", req.Mode, actor(r))
	}
	if req.KillSwitch != nil {
		s.Kill = *req.KillSwitch
		s.audit("AI_KILL_SWITCH", fmt.Sprint(s.Kill), actor(r))
	}

	jsonOut(w, map[string]any{"configuredMode": s.Mode, "effectiveMode": effectiveMode(), "killSwitch": s.Kill})
}

func asns(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	if r.Method == http.MethodGet {
		jsonOut(w, s.ASNs)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var x ASN
	if err := json.NewDecoder(r.Body).Decode(&x); err != nil || x.Supplier == "" || x.SKU == "" || x.Expected <= 0 {
		bad(w, "supplier, sku, expected required")
		return
	}

	x.ID = nextID("ASN")
	x.Status = "EXPECTED"
	x.CreatedAt = now()
	s.ASNs = append([]ASN{x}, s.ASNs...)
	s.audit("ASN_CREATED", x.ID+" "+x.SKU, actor(r))
	jsonOut(w, x)
}

func asnAction(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/asns/"), "/")
	if len(parts) != 2 || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	for i := range s.ASNs {
		a := &s.ASNs[i]
		if a.ID != parts[0] {
			continue
		}

		switch parts[1] {
		case "receive":
			var req struct {
				Qty int `json:"qty"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.Qty <= 0 || a.Received+req.Qty > a.Expected {
				bad(w, "invalid receive quantity")
				return
			}
			a.Received += req.Qty
			if a.Received == a.Expected {
				a.Status = "RECEIVED"
			} else {
				a.Status = "RECEIVING"
			}

			inv := s.Inv[a.SKU]
			if inv == nil {
				inv = &Inventory{SKU: a.SKU, Location: "RECEIVING"}
				s.Inv[a.SKU] = inv
			}
			inv.Receiving += req.Qty

			t := Task{ID: nextID("T"), Type: "PUTAWAY", Ref: a.ID, Status: "READY", Executor: "WCS"}
			s.Tasks = append([]Task{t}, s.Tasks...)
			s.audit("ASN_RECEIVED", fmt.Sprintf("%s +%d", a.ID, req.Qty), actor(r))
			jsonOut(w, a)
			return

		case "putaway":
			if a.Received <= 0 {
				bad(w, "nothing received")
				return
			}
			inv := s.Inv[a.SKU]
			if inv == nil || inv.Receiving < a.Received {
				bad(w, "receiving inventory inconsistent")
				return
			}
			inv.Receiving -= a.Received
			inv.Available += a.Received
			if inv.Location == "RECEIVING" {
				inv.Location = "A01-03-01"
			}
			a.Status = "PUTAWAY_DONE"
			completeTask(a.ID, "PUTAWAY")
			s.audit("PUTAWAY_COMPLETED", a.ID, actor(r))
			jsonOut(w, a)
			return
		}
	}

	http.NotFound(w, r)
}

func inventory(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	arr := make([]*Inventory, 0, len(s.Inv))
	for _, v := range s.Inv {
		arr = append(arr, v)
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].SKU < arr[j].SKU })
	jsonOut(w, arr)
}

func orders(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	if r.Method == http.MethodGet {
		jsonOut(w, s.Orders)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var o Order
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil || o.Customer == "" || o.SKU == "" || o.Qty <= 0 {
		bad(w, "customer, sku, qty required")
		return
	}

	o.ID = nextID("SO")
	o.Status = "CREATED"
	o.CreatedAt = now()
	s.Orders = append([]Order{o}, s.Orders...)
	s.audit("ORDER_CREATED", o.ID+" "+o.SKU, actor(r))
	jsonOut(w, o)
}

func orderAction(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/orders/"), "/")
	if len(parts) != 2 || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	id, act := parts[0], parts[1]
	for i := range s.Orders {
		o := &s.Orders[i]
		if o.ID != id {
			continue
		}
		inv := s.Inv[o.SKU]

		switch act {
		case "allocate":
			if o.Status != "CREATED" {
				bad(w, "order not CREATED")
				return
			}
			if inv == nil || inv.Available < o.Qty {
				bad(w, "insufficient stock")
				return
			}
			inv.Available -= o.Qty
			inv.Allocated += o.Qty
			o.Status = "ALLOCATED"
			s.Tasks = append([]Task{
				{ID: nextID("T"), Type: "PICK", Ref: o.ID, Status: "READY", Executor: "WCS"},
				{ID: nextID("T"), Type: "PACK", Ref: o.ID, Status: "WAITING", Executor: "SYSTEM"},
				{ID: nextID("T"), Type: "SHIP", Ref: o.ID, Status: "WAITING", Executor: "SYSTEM"},
			}, s.Tasks...)

		case "pick":
			if o.Status != "ALLOCATED" {
				bad(w, "order not ALLOCATED")
				return
			}
			inv.Allocated -= o.Qty
			inv.Picked += o.Qty
			o.Status = "PICKED"
			completeTask(o.ID, "PICK")
			readyTask(o.ID, "PACK")

		case "pack":
			if o.Status != "PICKED" {
				bad(w, "order not PICKED")
				return
			}
			o.Status = "PACKED"
			completeTask(o.ID, "PACK")
			readyTask(o.ID, "SHIP")

		case "ship":
			if o.Status != "PACKED" {
				bad(w, "order not PACKED")
				return
			}
			inv.Picked -= o.Qty
			o.Status = "SHIPPED"
			completeTask(o.ID, "SHIP")

		default:
			http.NotFound(w, r)
			return
		}

		s.audit("ORDER_"+strings.ToUpper(act), o.ID, actor(r))
		jsonOut(w, o)
		return
	}

	http.NotFound(w, r)
}

func completeTask(ref, typ string) {
	for i := range s.Tasks {
		if s.Tasks[i].Ref == ref && s.Tasks[i].Type == typ {
			s.Tasks[i].Status = "SUCCEEDED"
		}
	}
}

func readyTask(ref, typ string) {
	for i := range s.Tasks {
		if s.Tasks[i].Ref == ref && s.Tasks[i].Type == typ {
			s.Tasks[i].Status = "READY"
		}
	}
}

func tasks(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()
	jsonOut(w, s.Tasks)
}

func taskAction(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/tasks/"), "/")
	if len(parts) != 2 || parts[1] != "retry" || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	for i := range s.Tasks {
		if s.Tasks[i].ID != parts[0] {
			continue
		}
		if s.Tasks[i].Status != "FAILED" {
			bad(w, "only FAILED task can retry")
			return
		}
		s.Tasks[i].Status = "READY"
		s.audit("TASK_RETRY", s.Tasks[i].ID, actor(r))
		jsonOut(w, s.Tasks[i])
		return
	}
	http.NotFound(w, r)
}

func audit(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()
	jsonOut(w, s.Audits)
}

func aiPlan(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	if effectiveMode() != "AI_WAREHOUSE_OS" {
		bad(w, "AI Warehouse OS is not active")
		return
	}

	plan := []map[string]any{}
	for _, a := range s.ASNs {
		if a.Received > 0 && a.Status != "PUTAWAY_DONE" {
			plan = append(plan, map[string]any{
				"action": "PUTAWAY", "ref": a.ID,
				"reason": "receiving stock pending", "risk": "LOW",
			})
		}
	}
	for _, o := range s.Orders {
		if o.Status == "CREATED" {
			plan = append(plan, map[string]any{
				"action": "ALLOCATE", "ref": o.ID,
				"reason": "outbound order waiting allocation", "risk": "LOW",
			})
		}
	}
	if len(plan) == 0 {
		plan = append(plan, map[string]any{
			"action": "NOOP", "reason": "warehouse is balanced", "risk": "LOW",
		})
	}

	s.audit("AI_PLAN_GENERATED", fmt.Sprintf("%d suggestions", len(plan)), actor(r))
	jsonOut(w, map[string]any{"effectiveMode": effectiveMode(), "suggestions": plan})
}
