// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package auth

import (
	"context"
	"log"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	headerAuthorize = "authorization"
	authedMarker    = "some_context_marker"
)

// 从ctx中获取authorization对应的值(scheme+token)
// 如果scheme一致，返回token
func AuthFromMD(ctx context.Context, expectedScheme string) (string, error) {
	val := metautils.ExtractIncoming(ctx).Get(headerAuthorize)
	if val == "" {
		return "", status.Errorf(codes.Unauthenticated, "Request unauthenticated with "+expectedScheme)
	}
	splits := strings.SplitN(val, " ", 2)
	if len(splits) < 2 {
		return "", status.Errorf(codes.Unauthenticated, "Bad authorization string")
	}
	if !strings.EqualFold(splits[0], expectedScheme) {
		return "", status.Errorf(codes.Unauthenticated, "Request unauthenticated with "+expectedScheme+"  and  "+splits[0])
	}
	return splits[1], nil
}

// 认证函数
// 返回一个函数，该函数接受ctx,从ctx中获取expectedScheme对应的token,然后与expectedToken对比，如果scheme和token均与预期一致，则认证通过(errorw为nil)
func BuildDummyAuthFunction(expectedScheme string, expectedToken string) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := AuthFromMD(ctx, expectedScheme)
		if err != nil {
			log.Printf("there is no scheme:%v\n", expectedScheme)
			return nil, err
		}
		if token != expectedToken {
			log.Println("bad token")
			return nil, status.Errorf(codes.PermissionDenied, "buildDummyAuthFunction bad token")
		}
		return context.WithValue(ctx, authedMarker, "marker_exists"), nil
	}
}
