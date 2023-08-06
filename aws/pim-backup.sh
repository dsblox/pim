# This backup script is specific to AWS assuming an S3 bucket, a machine with 
#   appropriate IAM permissions enabled, and AWS CLI installed

# Create the backup file
TIME=$(date --utc "+%Y%m%d_%H%M%SZ")
BACKUP_FILE="pim_backup_${TIME}.pgdump"
DATABASE_NAME="pim"
docker exec pim_db_1 pg_dump -U postgres $DATABASE_NAME --format=custom > $BACKUP_FILE

# Second, copy file to AWS S3
S3_BUCKET=s3://blockshots-pim-backups
S3_TARGET=$S3_BUCKET/$BACKUP_FILE
echo "Copying $BACKUP_FILE to $S3_TARGET"
aws s3 cp $BACKUP_FILE $S3_TARGET

#verify the backup was uploaded correctly
echo "Backup completed for $DATABASE_NAME"
BACKUP_RESULT=$(aws s3 ls $S3_BUCKET | tail -n 1)
echo "Latest S3 backup: $BACKUP_RESULT"

#clean up and delete the local backup file
rm $BACKUP_FILE

