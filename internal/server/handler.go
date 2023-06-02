package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/app"
)

type Handler struct {
	app    app.RequestValidator
	router map[string]http.HandlerFunc
}

func newHandler(app app.RequestValidator) *Handler {
	return &Handler{
		app:    app,
		router: make(map[string]http.HandlerFunc),
	}
}

type Error struct {
	Err string `json:"error"`
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router["/try"] = h.TryAuth
	h.router["/whitelist"] = h.Whitelist
	h.router["/blacklist"] = h.Blacklist
	h.router["/reset"] = h.Reset

	if handler, ok := h.router[r.URL.Path]; ok {
		handler(w, r)
		return
	}
	http.NotFound(w, r)
}

func getReq(r *http.Request) (app.Request, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return app.Request{}, err
	}
	defer r.Body.Close()

	var ar app.Request
	if err = json.Unmarshal(b, &ar); err != nil {
		return app.Request{}, err
	}
	return ar, nil
}

func getNetwork(r *http.Request) (app.Network, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return app.Network{}, err
	}
	defer r.Body.Close()

	var an app.Network
	if err = json.Unmarshal(b, &an); err != nil {
		return app.Network{}, err
	}
	return an, nil
}

func sendError(w http.ResponseWriter, err error) {
	b, err := json.Marshal(Error{err.Error()})
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{
			"error": "%s"
		}
		`, err.Error()),
		))
		return
	}
	w.Write(b)
}

func (h Handler) TryAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	ar, err := getReq(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		b, err := json.Marshal(Error{err.Error()})
		if err != nil {
			w.Write([]byte(fmt.Sprintf(`{
				"error": "%s"
			}
			`, err.Error()),
			))
			return
		}
		w.Write(b)
		return
	}
	if ar.IP == "" || ar.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok, err := h.app.TryAuth(ctx, ar)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		b, err := json.Marshal(Error{err.Error()})
		if err != nil {
			w.Write([]byte(fmt.Sprintf(`{
				"error": "%s"
			}
			`, err.Error()),
			))
			return
		}
		w.Write(b)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusTooManyRequests)
	}
}

func (h Handler) Whitelist(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.addToWhite(w, r)
		return
	}
	if r.Method == http.MethodDelete {
		h.removeFromWhite(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (h Handler) addToWhite(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	an, err := getNetwork(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		sendError(w, err)
	}
	if err = h.app.AddToWhitelist(ctx, an); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendError(w, err)
	}
}

func (h Handler) removeFromWhite(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	an, err := getNetwork(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		sendError(w, err)
	}
	if err = h.app.RemoveFromWhitelist(ctx, an); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendError(w, err)
	}
}

func (h Handler) Blacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.addToBlack(w, r)
		return
	}
	if r.Method == http.MethodDelete {
		h.removeFromBlack(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (h Handler) addToBlack(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	an, err := getNetwork(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		sendError(w, err)
	}
	if err = h.app.AddToBlacklist(ctx, an); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendError(w, err)
	}
}

func (h Handler) removeFromBlack(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	an, err := getNetwork(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		sendError(w, err)
	}
	if err = h.app.RemoveFromBlacklist(ctx, an); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendError(w, err)
	}
}

func (h Handler) Reset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	ar, err := getReq(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		sendError(w, err)
	}

	if err = h.app.ResetBuckets(ctx, ar.Login, ar.IP); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendError(w, err)
	}
}
