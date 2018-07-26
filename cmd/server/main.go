package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/rodrwan/fakeprovider/logger"

	"github.com/ulule/limiter/drivers/middleware/stdlib"

	"github.com/google/uuid"
	apierror "github.com/rodrwan/fakeprovider/api-error"
	corsLib "github.com/rs/cors"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/store/memory"
)

const (
	minCreateProcessTime = 2
	maxCreateProcessTime = 10
	minLoadProcessTime   = 2
	maxLoadProcessTime   = 10
)

var (
	port  = flag.String("port", "8080", "Service port")
	token = flag.String("token", "fasdfadfa9fj987afsdf", "Token for authenticated endpointds")
)

func main() {
	flag.Parse()

	cards := make([]*card, 0)
	// This is where the router is useful, it allows us to declare methods that
	// this path will be valid for
	cards = append(cards, newCard(&user{
		FirstName: "louane",
		LastName:  "vidal",
		Email:     "louane.vidal@example.com",
	}))
	cards = append(cards, newCard(&user{
		FirstName: "noel",
		LastName:  "peixoto",
		Email:     "noel.peixoto@example.com",
	}))
	cards = append(cards, newCard(&user{
		FirstName: "manuel",
		LastName:  "lorenzo",
		Email:     "manuel.lorenzo@example.com",
	}))
	cards = append(cards, newCard(&user{
		FirstName: "alberto",
		LastName:  "lozano",
		Email:     "alberto.lozano@example.com",
	}))

	cc := &Context{
		cards: cards,
	}

	rate := limiter.Rate{
		Period: 10 * time.Second,
		Limit:  2,
	}
	store := memory.NewStore()

	// middlewares

	fakeLogger := logger.NewLogger("fakeprovider")
	rateLimitMid := stdlib.NewMiddleware(
		limiter.New(store, rate),
		stdlib.WithForwardHeader(true),
		stdlib.WithLimitReachedHandler(func(w http.ResponseWriter, r *http.Request) {
			apierror.NewError("Limit exceeded", http.StatusTooManyRequests).Write(w)
		}),
	)
	auth := NewAuthMiddleware(*token)

	r := NewRouter()
	r.GET("/", fakeLogger.Handle(rateLimitMid.Handler(ContextHandler{cc, getAllCardsHandler})))
	r.POST("/cards", fakeLogger.Handle(rateLimitMid.Handler(ContextHandler{cc, create})))
	r.POST("/load", fakeLogger.Handle(rateLimitMid.Handler(ContextHandler{cc, loadHandler})))
	r.PATCH("/cards/:id/info", fakeLogger.Handle(auth.Handle(ContextHandler{cc, patch})))

	// We can then pass our router (after declaring all our routes) to this method
	// (where previously, we were leaving the secodn argument as nil)
	log.Printf("server running on %s", fmt.Sprintf(":%s", *port))

	cors := corsLib.New(corsLib.Options{
		AllowedOrigins:     []string{"*"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "Credentials"},
		AllowedMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials:   true,
		OptionsPassthrough: true,
		Debug:              true,
	})

	// routes := negroni.Wrap(r)
	// n := negroni.New()
	// n.Use(negroni.NewRecovery())

	// logger := logrus.New()
	// logger.Formatter = &logrus.JSONFormatter{}

	// n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "fake-provider"))
	mux := http.NewServeMux()

	mux.Handle("/", cors.Handler(r))
	// n.UseHandler(mux)
	panic(http.ListenAndServe(fmt.Sprintf(":%s", *port), mux))
}

func unmarshalJSON(r io.ReadCloser, v interface{}) error {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return json.Unmarshal(body, v)
}

func randomProcessTime(min, max int) time.Duration {
	rand.Seed(time.Now().UTC().UnixNano())
	return time.Duration(rand.Intn(max-min) + min)
}

type response struct {
	Status int         `json:"-"`
	Data   interface{} `json:"data,omitempty"`
	Meta   interface{} `json:"meta,omitempty"`
}

// Write writes a ApplicationResposne to the given response writer encoded as JSON.
func (r *response) Write(w http.ResponseWriter) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status)
	_, err = w.Write(b)
	return err
}

type errorResponse struct {
	Status int           `json:"-"`
	Error  *errorMessage `json:"error,omitempty"`
}

type errorMessage struct {
	Message string `json:"message"`
}

// Write writes a ApplicationResposne to the given response writer encoded as JSON.
func (er *errorResponse) Write(w http.ResponseWriter) error {
	b, err := json.Marshal(er)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(er.Status)
	_, err = w.Write(b)
	return err
}

// NewError ...NewError
func NewError(msg string, status int) *errorResponse {
	return &errorResponse{
		Status: status,
		Error: &errorMessage{
			Message: msg,
		},
	}
}

type user struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type card struct {
	ID          string    `json:"id"`
	NameOnCard  string    `json:"name_on_card"`
	PAN         string    `json:"pan"`
	ReferenceID string    `json:"reference_id"`
	ExpDate     string    `json:"exp_date"`
	CVV         string    `json:"cvv"`
	Balance     int64     `json:"balance"`
	User        *user     `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (c *card) SetNameOnCard(u *user) {
	c.NameOnCard = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

func (c *card) SetPAN() {
	c.PAN = randomStringNumber(16)
}

func (c *card) SetReferenceID() {
	c.ReferenceID = randomStringNumber(8)
}

func (c *card) SetExpDate() {
	month := pickMonth()
	year := pickYear()
	c.ExpDate = fmt.Sprintf("%s/%s", month, year)
}

func (c *card) SetBalance(balance int64) {
	c.Balance = balance
}

func (c *card) SetUser(u *user) {
	c.User = u
}

func newCard(u *user) *card {
	c := &card{}

	c.ID = newID()
	c.SetNameOnCard(u)
	c.SetPAN()
	c.SetExpDate()
	c.SetReferenceID()
	c.SetBalance(0)
	c.SetUser(u)
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	processTime := randomProcessTime(minCreateProcessTime, maxCreateProcessTime) * time.Second
	log.Printf("Waiting for %.2fs", processTime.Seconds())
	time.Sleep(processTime)

	return c
}

func randomStringNumber(n int) string {
	rand.Seed(time.Now().UnixNano())
	var numbers = []rune("0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = numbers[rand.Intn(len(numbers))]
	}
	return string(b)
}

func pickMonth() string {
	rand.Seed(time.Now().UnixNano())
	var months = []string{
		"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12",
	}
	return months[rand.Intn(len(months))]
}

func pickYear() string {
	rand.Seed(time.Now().UnixNano())
	var years = []string{
		"19", "20", "21", "22", "23",
	}
	return years[rand.Intn(len(years))]
}

// newID creates a new UUID.
func newID() string {
	u2, err := uuid.NewRandom()
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return ""
	}
	return u2.String()
}
