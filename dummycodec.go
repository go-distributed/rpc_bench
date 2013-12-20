package rpc_bench

import (
	"net/rpc"
	"io"
	"encoding/binary"
)

type DummyCodec struct {
	rwc io.ReadWriteCloser
	b []byte
}

// server interfaces
func (d *DummyCodec) ReadRequestHeader(r *rpc.Request) error {
	r.ServiceMethod = "Foo.Dummy"
	_, err := d.rwc.Read(d.b)
	if err != nil {
		return err
	}
	seq, _ := binary.Uvarint(d.b)
	r.Seq = seq
	
	return nil
}

func (d *DummyCodec) ReadRequestBody(body interface{}) error {
	return nil
}

func (d *DummyCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	binary.PutUvarint(d.b, r.Seq)
	_, err = d.rwc.Write(d.b)
	if err != nil {
		return err
	}
	return nil
}

// client interfaces
func (d *DummyCodec) WriteRequest(r *rpc.Request, body interface{}) error {
	binary.PutUvarint(d.b, r.Seq)
	_, err := d.rwc.Write(d.b)
	if err != nil {
		return err
	}
	return nil
}

func (d *DummyCodec) ReadResponseHeader(r *rpc.Response) error {
	_, err := d.rwc.Read(d.b)
	if err != nil {
		return err
	}
	seq, _ := binary.Uvarint(d.b)
	r.Seq = seq
	
	return nil
}

func (d *DummyCodec) ReadResponseBody(body interface{}) error {
	return nil
}

func (d *DummyCodec) Close() error {
	return d.rwc.Close()
}