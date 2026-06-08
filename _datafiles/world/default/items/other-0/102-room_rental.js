

/**
 * Called when a user purchases the item from a shop.
 * @param {ActorObject} user - The user purchasing the item.
 * @param {ItemObject} item - The item being purchased.
 * @param {RoomObject} room - The room where the purchase occurred.
 * @returns {boolean} Return false to allow the default purchase behavior.
 */
function onPurchase(user, item, room) {

    var newRoomIds = CreateInstancesFromRoomIds( [432] );

    if ( newRoomIds[432] ) {

        SendUserMessage(user.UserId(), "You are directed to a room upstairs with a large bed. How inviting...");
        SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" says something to the Inn keeper and is escorted to another room.", user.UserId());
        
        user.MoveRoom(newRoomIds[432]);
        return false;
    } 

    return false;
}

