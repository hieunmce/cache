package resource

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestFetchCache_Fetch_MultipleID_NonBlock(t *testing.T) {
	var (
		ids = []string{
			"7e1588ab-2bf9-44b0-a7ae-41c9ef3c86af",
			"6bff5837-c618-46ac-ae27-4544e08b099e",
			"ecdcb84d-7c42-4242-9e46-bb3b9fbe4312",
			"530bc2d6-3023-4206-aa0c-9d21dbb7d0a9",
			"f76dd77f-a46d-45d9-a0cd-e2fe66a6820a",
			"896ad76d-0022-4543-be16-91d68f747715",
			"d90f89ae-cc4c-45b4-9b58-8ff322a39f5b",
			"3b0dcd0b-e585-43be-a5fe-aba5370bfa77",
			"4e63323c-844e-4d85-8cc4-93bb2830aeb5",
			"c5bcb00e-1141-4a17-93e2-b91c7d42cdcc",
			"8e41c031-843a-453c-8a80-ff03c7eb265c",
		}
	)

	sleepDuration := 10 * time.Millisecond

	mockedFetcher := &FetcherMock{
		FetchFunc: func(ctx context.Context, id string) (*Model, error) {
			time.Sleep(sleepDuration)
			return &Model{Name: id}, nil
		},
	}
	fc := NewCache(mockedFetcher)
	var wg sync.WaitGroup
	wg.Add(len(ids))
	start := time.Now()

	for i := 0; i < len(ids); i++ {
		go func(ii int) {
			_, _ = fc.Fetch(context.Background(), ids[ii])
			wg.Done()
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)
	if elapsed > sleepDuration*2 {
		t.Errorf("FetchCache.Fetch() expect duration for 11 goroutine smaller than %v, have duration for the call %v", sleepDuration*2, elapsed)
	}
}

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
			callCount:        1000,
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
