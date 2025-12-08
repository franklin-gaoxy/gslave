
## backup

```
mkdir -p /tmp/mongo-backup && \
mongodump \
  --uri "mongodb://%2Ftmp%2Fmongodb-27017.sock" \
  --username root \
  --password 'Mongo@123456' \
  --authenticationDatabase admin \
  --db gslave \
  --archive=/tmp/mongo-backup/gslave-$(date +%F).gz \
  --gzip \
  --verbose
```


## restore

```
mongorestore --gzip --archive=/tmp/mongo-backup/gslave-2025-12-08.gz
```
