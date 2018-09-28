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
		fakeFetchID = "dca76878-a8f6-4ff5-b263-1e8c7e61bc20"
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
			name: "test normal case",
			args: args{
				ctx: context.Background(),
				id:  fakeFetchID,
			},
			want:             &Model{Name: "lorem"},
			serviceCallCount: 1,
			callCount:        1000,
		},
		{
			name: "test normal case",
			args: args{
				ctx: context.Background(),
				id:  "eeebf00e-3401-40e8-b912-c3b64e429e5c",
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
					t.Errorf("FetchCache.Fetch() expect service count = %v, want service call count %v", tt.serviceCallCount, serviceCallCount)
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
