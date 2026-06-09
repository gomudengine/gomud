
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You catch a whiff of a strange odor carried by the smoke.' );
    SendRoomMessage(actor.GetRoomId(),  'You notice the eyes of '+actor.GetCharacterName(true)+' glaze oover.', actor.UserId());
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You find yourself in a blissful haze, wanting nothing more than to sit and watch the world go by.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' looks blissfully content.', actor.UserId());
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You snap out of your haze.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' shakes off the haze and the glaze in their eyes fades away.', actor.UserId());
}
