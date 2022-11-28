package interceptor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	scheme = "good_scheme"
	token  = "some_good_token"
)

func ctxWithToken(ctx context.Context, scheme string, token string) context.Context {
	md := metadata.Pairs("authorization", fmt.Sprintf("%s %v", scheme, token))
	nCtx := metautils.NiceMD(md).ToOutgoing(ctx)
	return nCtx
}

func UnaryClientInterceptor1(ctx context.Context, method string, req, rsp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {

	start := time.Now()
	defer func() {
		in, _ := json.Marshal(req)
		out, _ := json.Marshal(rsp)
		inStr, outStr := string(in), string(out)
		duration := int64(time.Since(start) / time.Millisecond)

		log.Println("grpc1", method, "in1:", inStr, "out1:", outStr, "duration/ms1", duration)

	}()
	ctx = ctxWithToken(ctx, scheme, token)
	return invoker(ctx, method, req, rsp, cc, opts...)
}

func UnaryClientInterceptor2(ctx context.Context, method string, req, rsp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {

	start := time.Now()
	defer func() {
		in, _ := json.Marshal(req)
		out, _ := json.Marshal(rsp)
		inStr, outStr := string(in), string(out)
		duration := int64(time.Since(start) / time.Millisecond)

		log.Println("grpc2", method, "in2:", inStr, "out2:", outStr, "duration/ms2", duration)

	}()
	return invoker(ctx, method, req, rsp, cc, opts...)
}
