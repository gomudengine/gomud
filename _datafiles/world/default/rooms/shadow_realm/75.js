

/**
 * Called each round when the room is idle (no players or on a timer).
 * @param {RoomObject} room - The room.
 * @returns {boolean} Return true if the event was handled.
 */
function onIdle(room) {
    
    if ( room.AddTemporaryExit('shimmering portal', ':cyan', 0, '15 minutes') ) {
        room.SendText('A portal to the world of the living appears!');
    }
    return false;
}

/**
 * Called when a user exits the room.
 * @param {ActorObject} user - The user exiting the room.
 * @param {RoomObject} room - The room being exited.
 * @returns {boolean} Return true if the event was handled.
 */
function onExit(user , room) {
    // Remove the healing buff if they are leaving
    user.RemoveBuff(24);
}
