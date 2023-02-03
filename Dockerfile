FROM gcr.io/prysmaticlabs/build-agent AS builder

WORKDIR /workspace

COPY . /workspace/.

RUN bazel build cmd/beacon-chain:beacon-chain --config=release

FROM ubuntu

COPY --from=builder /workspace/bazel-bin/cmd/beacon-chain/beacon-chain_/beacon-chain .

ENTRYPOINT ["./beacon-chain"]
