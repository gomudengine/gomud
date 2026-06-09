
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You notice that the world looks more vibrant and alive.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' seems to be distracted.', actor.UserId());
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'The colors and sounds of the world captivate you.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' stares into the distance.', actor.UserId());
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'The world starts to look a little more normal.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' seems sober again.', actor.UserId());
}
