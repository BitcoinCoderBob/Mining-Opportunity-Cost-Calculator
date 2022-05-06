package miningprofitability

import (
	"encoding/json"
	"io"
	"net/http"

	"Mining-Profitability/pkg/appcontext"
)

type RequestPayload struct {
	Token              string  `json:"token"`
	StartDate          string  `json:"startDate"`
	KwhPrice           int     `json:"kwhPrice"`
	Watts              int     `json:"watts"`
	ElectricCosts      float64 `json:"electicCosts"`
	UptimePercent      int     `json:"updtimePercent"`
	FixedCosts         float64 `json:"fixedCosts"`
	BitcoinMined       float64 `json:"bitcoinMined"`
	MessariApiKey      string  `json:"messariApiKey"`
	HideBitcoinOnGraph bool    `json:"hideBitcoinOnGraph"`
}

type Handler struct {
	actx *appcontext.AppContext
}

func NewHandler(actx *appcontext.AppContext) *Handler {
	return &Handler{actx}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.actx.Logger.Debug("endpoint only accepts POST")
		w.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	_, err := io.ReadAll(r.Body)
	if err != nil {
		h.actx.Logger.WithError(err).Error("error reading the request body")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("error reading body"))

		return
	}

	var requestPayload RequestPayload
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&requestPayload); err != nil {
		h.actx.Logger.WithError(err).Error("error parsing the request body into requestpayload struct")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("error unmarshaling body"))

		return
	}

	h.handleRequest(w, nil)
}

func (h *Handler) handleRequest(w http.ResponseWriter, requestPayload *RequestPayload) {
	//TODO: handle post request to calculate here
}
