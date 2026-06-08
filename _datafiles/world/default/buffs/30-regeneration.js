
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'Your body\'s natural healing feels super charged.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' begins to regenerate.', actor.UserId());
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 3));
    if (healAmt > 0) {
        SendUserMessage(actor.UserId(), 'You regenerate for <ansi fg="healing">'+String(healAmt)+' damage</ansi>!');
    }
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'Your enhanced regeneration goes away.');
}
