#!/bin/bash
protoc --go_out=plugins=grpc:. greet/proto/greet.proto
protoc --go_out=plugins=grpc:. calculator/proto/calculator.proto
protoc --go_out=plugins=grpc:. blog/proto/blog.proto

