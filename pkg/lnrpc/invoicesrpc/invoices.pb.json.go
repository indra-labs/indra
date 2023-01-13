// Code generated by falafel 0.9.1. DO NOT EDIT.
// source: invoices.proto

package invoicesrpc

import (
	"context"

	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func RegisterInvoicesJSONCallbacks(registry map[string]func(ctx context.Context,
	conn *grpc.ClientConn, reqJSON string, callback func(string, error))) {

	marshaler := &gateway.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
	}

	registry["invoicesrpc.Invoices.SubscribeSingleInvoice"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &SubscribeSingleInvoiceRequest{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewInvoicesClient(conn)
		stream, err := client.SubscribeSingleInvoice(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		go func() {
			for {
				select {
				case <-stream.Context().Done():
					callback("", stream.Context().Err())
					return
				default:
				}

				resp, err := stream.Recv()
				if err != nil {
					callback("", err)
					return
				}

				respBytes, err := marshaler.Marshal(resp)
				if err != nil {
					callback("", err)
					return
				}
				callback(string(respBytes), nil)
			}
		}()
	}

	registry["invoicesrpc.Invoices.CancelInvoice"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &CancelInvoiceMsg{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewInvoicesClient(conn)
		resp, err := client.CancelInvoice(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			callback("", err)
			return
		}
		callback(string(respBytes), nil)
	}

	registry["invoicesrpc.Invoices.AddHoldInvoice"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &AddHoldInvoiceRequest{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewInvoicesClient(conn)
		resp, err := client.AddHoldInvoice(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			callback("", err)
			return
		}
		callback(string(respBytes), nil)
	}

	registry["invoicesrpc.Invoices.SettleInvoice"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &SettleInvoiceMsg{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewInvoicesClient(conn)
		resp, err := client.SettleInvoice(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			callback("", err)
			return
		}
		callback(string(respBytes), nil)
	}

	registry["invoicesrpc.Invoices.LookupInvoiceV2"] = func(ctx context.Context,
		conn *grpc.ClientConn, reqJSON string, callback func(string, error)) {

		req := &LookupInvoiceMsg{}
		err := marshaler.Unmarshal([]byte(reqJSON), req)
		if err != nil {
			callback("", err)
			return
		}

		client := NewInvoicesClient(conn)
		resp, err := client.LookupInvoiceV2(ctx, req)
		if err != nil {
			callback("", err)
			return
		}

		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			callback("", err)
			return
		}
		callback(string(respBytes), nil)
	}
}
