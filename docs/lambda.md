# Lambda S3 -> S3

## Purpose

This Lambda ![function](./../cmd/main.go) listens on `Bucket_A` for uploads of `.geojson` files. These are typically files meeting the geojson spec for a `FeatureCollection`.

For each `Feature` contained in a `FeatureCollection` file, this function uses the [Viswalinham-Whyatt Algorithm](https://en.wikipedia.org/wiki/Visvalingam%E2%80%93Whyatt_algorithm) to priority rank the points in the shape, and save the result to `Bucket_B`. This function also saves a metadata file to `Bucket_B/meta` that contains the name, hash, and filepath of the feature.

## Deploying Function to Lambda

This function is **updated** as part of the repository CI. This CI assumes there is already an existing function to update.

## Frequently Used Commands + Reference

Update/Deploy Function:

```bash
GOOS=linux go build -o ./build/${FUNCTION} ./cmd/ &&\
    cd ./build && zip ./function.zip ${FUNCTION} &&\
    aws lambda update-function-code \
        --function-name ${FUNCTION} \
        --zip-file fileb://function.zip
```

Expected Env Vars:

```bash
S3_SHAPES_DEFAULT_REGION = us-east-1
S3_SHAPES_SRC_BUCKET = `Bucket_A`
S3_SHAPES_TARGET_BUCKET = `Bucket_B`
S3_WORKER_CONCURRENCY = 10
```
