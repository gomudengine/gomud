#!/usr/bin/env sh
#
# Helper script to compile/copy binary to test server on raspberry pi
#

RPI_HOST="mud@mud.local"
RPI_HOST_PATH="/home/mud"
RPI_BIN="go-mud-server-rpi"

# Build the raspberry pi binary... building on the rpi is problematic.
echo "Building bin for Raspberry Pi Zero"
make build_rpi_zero2w

# Kill the process before overwriting the binary
echo "Killing process on RaspPi: ${RPI_HOST}"
ssh "${RPI_HOST}" 'sudo pkill mud-server'

# Copy the binary over, delete the local binary, run the server again using the script setup on the server
echo "Copying bin to RaspPi: ${RPI_HOST}"
if ! scp "./${RPI_BIN}" "${RPI_HOST}:${RPI_HOST_PATH}/mud/${RPI_BIN}"; then
	exit 1
fi

echo "Starting Server on RaspPi: ${RPI_HOST}"
if ! rm "${RPI_BIN}"; then
	exit 1
fi
ssh "${RPI_HOST}" -f "cd ${RPI_HOST_PATH}; sudo ./startup-run-mud.sh"
