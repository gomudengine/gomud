
/**
 * Called when a user loses the item.
 * @param {ActorObject} user - The user who lost the item.
 * @param {ItemObject} item - The item.
 * @param {RoomObject} room - The room where the loss occurred.
 * @returns {void}
 */
function onLost(user, item, room) {
    SendUserMessage(user.UserId(), "You feel disappointment at the loss.");
}

/**
 * Called when a user finds or picks up the item.
 * @param {ActorObject} user - The user who found the item.
 * @param {ItemObject} item - The item.
 * @param {RoomObject} room - The room where the find occurred.
 * @returns {void}
 */
function onFound(user, item, room) {
    SendUserMessage(user.UserId(), "This feels... important.");
}
