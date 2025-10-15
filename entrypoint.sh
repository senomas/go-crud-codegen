#!/bin/sh
set -e

# Default UID/GID if not provided.
# In a CI/CD environment, you might want to default to a non-root user
# that is known to exist in the image.
USER_ID=${LOCAL_USER_ID:-9001}
GROUP_ID=${LOCAL_GROUP_ID:-9001}

echo "Starting with UID : $USER_ID, GID: $GROUP_ID"

# Create a group and user with the specified IDs.
# The '-o' flag allows reusing an existing GID/UID.
groupadd -g $GROUP_ID -o appgroup
useradd --shell /bin/bash -u $USER_ID -g $GROUP_ID -o -c "" -m appuser

# Set ownership of the user's home directory.
export HOME=/home/appuser
chown -R appuser:appgroup $HOME

chown -R appuser:appgroup /work

# Execute the command passed into the entrypoint (the Docker CMD)
# as the newly created user.
exec su appuser -c "$*"
