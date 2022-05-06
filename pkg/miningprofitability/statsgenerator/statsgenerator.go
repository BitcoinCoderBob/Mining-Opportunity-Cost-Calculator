package statsgenerator

import (
	"encoding/json"
	"io"
	"net/http"

	"Mining-Profitability/pkg/appcontext"
	"Mining-Profitability/pkg/calc"
)

type Handler struct {
	actx *appcontext.AppContext
}

func NewDataHandler(actx *appcontext.AppContext) *Handler {
	return &Handler{actx}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.actx.Logger.Debug("endpoint only accepts POST")
		w.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	a, err := io.ReadAll(r.Body)
	if err != nil {
		h.actx.Logger.WithError(err).Error("error reading the request body")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("error reading body"))

		return
	}

	var requestPayload calc.RequestPayload
	if err := json.Unmarshal(a, &requestPayload); err != nil {
		h.actx.Logger.WithError(err).Error("error parsing the request body into requestpayload struct")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("error unmarshaling body"))

		return
	}

	h.handleRequest(w, &requestPayload)
}

func (h *Handler) handleRequest(w http.ResponseWriter, requestPayload *calc.RequestPayload) {

	if requestPayload.SlushToken == nil && requestPayload.BitcoinMined == 0 {
		h.actx.Logger.Error("error must send either slush api token or bitcoinMined")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("error must send either slush api token or bitcoinMined"))

	}

	stats, err := h.actx.Calc.GenerateStats(*requestPayload, h.actx.ExternalData, h.actx.Utils)
	if err != nil {
		h.actx.Logger.WithError(err).Error("error must send either slush api token or bitcoinMined")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
	}
	if !requestPayload.ShowStrategyData {
		stats.AhData = make([]float64, 0)
		stats.AntiHomeMinerData = make([]float64, 0)
		stats.DcaData = make([]float64, 0)
	}

	byteRes, err := json.Marshal(stats)
	if err != nil {
		h.actx.Logger.WithError(err).Error("error must send either slush api token or bitcoinMined")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(byteRes)

}
