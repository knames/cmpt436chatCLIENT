package main

import (
  "errors"
  "sync"
)

var curClients map[string]*Client = make(map[string]*Client)
var curClientsMutex sync.RWMutex

type Client struct {
  Token string
  Name string
  CRoom *CRoom
  outMsg chan string
  Mutex sync.RWMutex
}

func NewClient(tok string) *Client {
  return &Client{
    Token: tok,
    Name: "Anon",
    CRoom: nil,
    outMsg: make(chan string),
  }
}

func AddClient(client *Client) error {
  curClientsMutex.Lock()
  defer curClientsMutex.Unlock()
  
  oClient := curClients[client.Token]
  if oClient != nil {
    return errors.New(ERR_TOK)
  }
  curClients[client.Token] = client
  return nil
}

func RemoveClient (tok string) error {
  curClientsMutex.Lock()
  defer curClientsMutex.Unlock()
  
  client := curClients[tok]
  if client == nil {
    return errors.New(ERR_TOK)
  }
  delete(curClients, tok)
  return nil
}

func GetClient(tok string) (*Client, error){
  curClientsMutex.Lock()
  defer curClientsMutex.Unlock()
  
  client := curClients[tok]
  if client == nil {
    return nil, errors.New(ERR_TOK)
  }
  return client, nil
}