package resource

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestFetchCache_Fetch(t *testing.T) {
	var (
		fakeFetchID     = "dca76878-a8f6-4ff5-b263-1e8c7e61bc20"
		notExistModelID = "5634aeed-2106-43de-ab7d-c0ad4b1e195e"
	)

	type fields struct {
		f Fetcher
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		want             *Model
		wantErr          bool
		serviceCallCount int
		callCount        int
	}{
		{
			name: "success get cache and check service number of request to service",
			args: args{
				ctx: context.Background(),
				id:  fakeFetchID,
			},
			want:             &Model{Name: "lorem"},
			serviceCallCount: 1,
			callCount:        100000,
		},
		{
			name: "failed get case by id not exist",
			args: args{
				ctx: context.Background(),
				id:  notExistModelID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCallCount := 0
			// make and configure a mocked Fetcher
			mockedFetcher := &FetcherMock{
				FetchFunc: func(ctx context.Context, id string) (*Model, error) {
					serviceCallCount++
					if id == fakeFetchID {
						return &Model{Name: "lorem"}, nil
					}

					return nil, errors.New("not found model")
				},
			}
			fc := NewCache(mockedFetcher)

			// test call count
			if tt.callCount > 0 {
				var wg sync.WaitGroup
				wg.Add(tt.callCount)
				for i := 0; i < tt.callCount; i++ {
					go func() {
						_, _ = fc.Fetch(tt.args.ctx, tt.args.id)
						wg.Done()
					}()
				}
				wg.Wait()
				if tt.serviceCallCount != serviceCallCount {
					t.Errorf("FetchCache.Fetch() expect service count = %v, have service call count %v", tt.serviceCallCount, serviceCallCount)
					return
				}
			}

			// test result
			got, err := fc.Fetch(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchCache.Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchCache.Fetch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchCache_Clear(t *testing.T) {
	var (
		fakeFetchID = "dca76878-a8f6-4ff5-b263-1e8c7e61bc20"
	)

	mockedFetcher := &FetcherMock{
		FetchFunc: func(ctx context.Context, id string) (*Model, error) {
			if id == fakeFetchID {
				return &Model{Name: "lorem"}, nil
			}

			return nil, errors.New("not found model")
		},
	}

	type fields struct {
		f Fetcher
	}
	type args struct {
		id string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		remainCount int
	}{
		{
			name: "success remove item by correct id",
			args: args{
				id: fakeFetchID,
			},
			remainCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewCache(mockedFetcher)
			_, _ = fc.Fetch(context.Background(), fakeFetchID)
			fc.Clear(tt.args.id)

			if len(fc.items) != tt.remainCount {
				t.Errorf("FetchCache.Clear() expect remain items count = %v, actual item count = %v", tt.remainCount, len(fc.items))
			}
		})
	}
}
