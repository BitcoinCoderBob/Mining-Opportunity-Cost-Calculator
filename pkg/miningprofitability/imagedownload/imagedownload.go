package imagedownload

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"Mining-Profitability/pkg/appcontext"
	"Mining-Profitability/pkg/calc"
)

type Handler struct {
	actx *appcontext.AppContext
}

func NewImageHandler(actx *appcontext.AppContext) *Handler {
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

	fileName, err := h.actx.Calc.GenerateImage(*requestPayload, h.actx.ExternalData, h.actx.Utils)
	if err != nil {
		h.actx.Logger.WithError(err).Error("error must send either slush api token or bitcoinMined")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
	}
	fn := filepath.Base(*fileName)
	file, err := os.OpenFile(*fileName, os.O_RDWR, 0644)
	if err != nil {
		h.actx.Logger.WithError(err).Error("error reading generated file")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fn))
	io.Copy(w, file)
	// defer os.Remove(*fileName)
	file.Close()
}
