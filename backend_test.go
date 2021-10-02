package main

import (
	"reflect"
	"testing"
)

func TestBackend_get(t *testing.T) {
	type fields struct {
		dir string
	}
	type args struct {
		tfID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			got, err := b.get(tt.args.tfID)
			if (err != nil) != tt.wantErr {
				t.Errorf("get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackend_getTfstateFilename(t *testing.T) {
	type fields struct {
		dir string
	}
	type args struct {
		tfID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			if got := b.getTfstateFilename(tt.args.tfID); got != tt.want {
				t.Errorf("getTfstateFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackend_lock(t *testing.T) {
	type fields struct {
		dir string
	}
	type args struct {
		tfID string
		lock []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			got, err := b.lock(tt.args.tfID, tt.args.lock)
			if (err != nil) != tt.wantErr {
				t.Errorf("lock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackend_pruge(t *testing.T) {
	type fields struct {
		dir string
	}
	type args struct {
		tfID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			if err := b.pruge(tt.args.tfID); (err != nil) != tt.wantErr {
				t.Errorf("pruge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackend_unlock(t *testing.T) {
	type fields struct {
		dir string
	}
	type args struct {
		tfID string
		lock []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			if err := b.unlock(tt.args.tfID, tt.args.lock); (err != nil) != tt.wantErr {
				t.Errorf("unlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackend_update(t *testing.T) {
	type fields struct {
		dir string
	}
	type args struct {
		tfID    string
		tfstate []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				dir: tt.fields.dir,
			}
			if err := b.update(tt.args.tfID, tt.args.tfstate); (err != nil) != tt.wantErr {
				t.Errorf("update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConflictError_Error(t *testing.T) {
	type fields struct {
		StatusCode int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConflictError{
				StatusCode: tt.fields.StatusCode,
			}
			if got := c.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileNotExistsError_Error(t *testing.T) {
	type fields struct {
		Info string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &FileNotExistsError{
				Info: tt.fields.Info,
			}
			if got := n.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
