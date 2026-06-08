
/**
 * Called each round when the room is idle (no players or on a timer).
 * @param {RoomObject} room - The room.
 * @returns {boolean} Return true if the event was handled.
 */
function onIdle(room) {

    if ( UtilGetRoundNumber()%30 == 0 ) {
        SendRoomMessage(room.RoomId(), "A huge wave crashes against the shore, but as it receeds, you notice a small path of shallow water you can follow to a large rock island.");
        room.AddTemporaryExit("shallow water", "shallow water", 828, 10);
    }

    return false;
}