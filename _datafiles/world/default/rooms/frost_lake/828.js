



lastSpawnRound = 0;

/**
 * Called when a user enters the room.
 * @param {ActorObject} user - The user entering the room.
 * @param {RoomObject} room - The room being entered.
 * @returns {boolean} Return false to suppress the automatic look.
 */
function onEnter(user, room) {
    roundNow = UtilGetRoundNumber();
    nextSpawnRound = lastSpawnRound + UtilGetSecondsToRounds(30);
    if ( lastSpawnRound > 0 && roundNow < nextSpawnRound ) {
        return true;
    }

    allItems = room.GetItems();

    oarExists = false;
    for ( i=0; i<allItems.length; i++ ) {
        if ( allItems[i].ItemId() == 10016 ) {
            oarExists = true;
            return true;
        }
    }

    if ( !oarExists ) {
        room.SpawnItem(10016, false);
        lastSpawnRound = roundNow;
    }

    return true;
}


