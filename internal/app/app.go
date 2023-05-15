package app

import "context"

type Request struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

type Network struct {
	IP string `json:"network"`
}

type RequestValidator interface {
	TryAuth(context.Context, Request) (bool, error)
	ResetBuckets(context.Context, string, string) error
	AddToBlacklist(context.Context, Network) error
	RemoveFromBlacklist(context.Context, Network) error
	AddToWhitelist(context.Context, Network) error
	RemoveFromWhitelist(context.Context, Network) error
}
