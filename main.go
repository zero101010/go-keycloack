package main

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"log"
	http "net/http"
	oidc "github.com/coreos/go-oidc"
	)

var (
	clientID = "app"
	clientSecret = "62dacf64-5c7f-494f-a745-167f5f5cab59"
	)

func main(){
	// CONFIG KEYCLOAK
	ctx:= context.Background()
	provider, error := oidc.NewProvider(ctx, "http://localhost:8080/auth/realms/first-realm")
	if error!= nil{
		log.Fatal(error)
	}
	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://localhost:8081/auth/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "roles"},
	}
	state:= "magica"

	// CRIAR ENDPOINTS

	// ENDPOINT DE REDIRECIONAMENTO PARA
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer,request,config.AuthCodeURL(state),http.StatusFound)
	})

	http.HandleFunc("/auth/callback", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("state") != state{
			http.Error(writer,"O estado da aplicação não é o mesmo que esperamos", http.StatusBadRequest)
			return
		}
		oauth2Token, error := config.Exchange(ctx,request.URL.Query().Get("code"))
		if error !=nil {
			http.Error(writer,"Erro para pegar o token", http.StatusBadRequest)
			return
		}
		rawIDToken, err := oauth2Token.Extra("id_token").(string)
		if !err{
			http.Error(writer,"Não achamos o id_token", http.StatusBadRequest)
			return
		}

		resp:=
			struct {
				OAuth2Token *oauth2.Token
				RawIDTOKEN string
			}{
				oauth2Token,rawIDToken,
			}

		data , errJson := json.MarshalIndent(resp,"","	")
		if errJson!=nil{
			http.Error(writer,"Ocorreu um erro ao tentar transformar em json", http.StatusBadRequest)
			return
		}

		writer.Write(data)

	})


	// SERVER QUE ESCUTARÁ AS REQUISIÇÕES
	log.Fatal(http.ListenAndServe(":8081", nil))
}