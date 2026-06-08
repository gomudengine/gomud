
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'re on the ground.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is on the ground.', actor.UserId());
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'re trying to get up.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is getting up.', actor.UserId());
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'re standing again.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is standing again.', actor.UserId());
}
