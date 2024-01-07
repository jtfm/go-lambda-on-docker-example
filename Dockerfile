FROM public.ecr.aws/lambda/provided:al2-arm64
COPY main ${LAMBDA_TASK_ROOT}/main
ENTRYPOINT ["./main"]