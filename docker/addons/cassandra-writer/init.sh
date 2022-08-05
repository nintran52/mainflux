docker-compose -f docker/addons/cassandra-writer/docker-compose.yml --env-file docker/.env up -d
sleep 20
docker exec mainfluxlabs-cassandra cqlsh -e "CREATE KEYSPACE IF NOT EXISTS mainflux WITH replication = {'class':'SimpleStrategy','replication_factor':'1'};"
