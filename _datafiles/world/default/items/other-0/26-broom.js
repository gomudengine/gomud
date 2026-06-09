
/**
 * Called when a user issues a sweep command on the item.
 * @param {ActorObject} user - The user issuing the command.
 * @param {ItemObject} item - The item.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_sweep(user, item, room) {

    SendUserMessage(user.UserId(), "You sweep the floors thoroughly, until not a single dust bunny can be found.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" sweeps their heart out with <ansi fg=\"item\">"+item.Name(true)+"</ansi>.", user.UserId());   

    room.RemoveMutator("dusty-floors");

    return true;
}

