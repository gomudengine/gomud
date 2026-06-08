
lastSpawnRound = 0;

/**
 * Called when a user enters the room.
 * @param {ActorObject} user - The user entering the room.
 * @param {RoomObject} room - The room being entered.
 * @returns {boolean} Return false to suppress the automatic look.
 */
function onEnter(user, room) {

    if ( !user.HasQuest("6-return") ) {
        room.RepeatSpawnItem(10, 30);
    }
 
    return true;
}
