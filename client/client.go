package client

import (
	"bytes"
	"context"
	"io"
	"net"

	"github.com/tidwall/resp"
)

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (c *Client) Set(ctx context.Context, key string, val string) error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	wr := resp.NewWriter(buf)
	wr.WriteArray([]resp.Value{resp.StringValue("SET"), resp.StringValue(key), resp.StringValue(val)})
	// _, err = conn.Write(buf.Bytes())
	_, err = io.Copy(conn, buf)
	return err

}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	wr := resp.NewWriter(buf)
	wr.WriteArray([]resp.Value{resp.StringValue("GET"), resp.StringValue(key)})
	// _, err = conn.Write(buf.Bytes())
	_, err = io.Copy(conn, buf)
	if err != nil {
		return "", err
	}
	b := make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		return "", nil
	}
	return string(b[:n]), nil

}
