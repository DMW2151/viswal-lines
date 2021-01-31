# ElasticSearch

## Deploying `ElasticSearch` on EC2

The Elasticsearch "Cluster" is deployed to run in single node mode on an EC2 instance. Instructions below specific to constraints on `t2.Micro` w. 1G RAM, 8G storage.

- Install Elasticsearch on the instance. See section 1 - [Installing Elasticsearch](https://www.digitalocean.com/community/tutorials/how-to-install-and-configure-elasticsearch-on-ubuntu-18-04)
  
- By default, Elasticsearch will attempt to start w. 1G and will throw `out of memory` on startup. Modify `/etc/elasticsearch/jvm.options` to allocate 256M on startup `-Xms256m`.
  
- By default, Elasticsearch will only listen on `localhost`. Modify the following in `/etc/elasticsearch/elasticsearch.yml` to listen for all traffic.
    - `network.host` to `0.0.0.0`
    - `http.port` to `9200`
    - `discovery.seed_hosts` to `["host1", "host2"]`  
  
- See [StackOverflow](https://stackoverflow.com/questions/20105448/access-ec2-port-9200-from-external-service) thread on accessing port 9200 from external service for more detail.

## Frequently Used Commands + Reference

Create an Index with autocomplete on

```bash
curl -X PUT "localhost:9200/shapes?pretty" \
    -H 'Content-Type: application/json' \
    -d '{ "mappings": {
            "properties": {
                "Name": {"type": "completion"}
            }
        }
    }'
```

Sample Query

```bash
curl -X POST " http://localhost:9200/shapes/_search?pretty&pretty" \
        -H 'Content-Type: application/json' \
        -d '{  "suggest": {
                "name-suggest": {
                    "prefix": "Bal",
                    "completion": {
                        "field": "Name"  
                    }
                }
            }
        }'
```

Redeploy - built local; upload binary

```bash
GOOS=linux go build -o ./api/main ./api/ && 
rsync -ave "ssh -i ${KEYPAIR_FILE}" api/main ${ELASTIC_SERVER}:/home/ubuntu/
```
