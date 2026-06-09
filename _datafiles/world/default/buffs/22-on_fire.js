
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You catch on <ansi fg="red">fire</ansi>!');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' caught on <ansi fg="red">fire</ansi>!', actor.UserId());
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    dmgAmt = Math.abs(actor.AddHealth(-1*UtilDiceRoll(2, 6)));

    SendUserMessage(actor.UserId(),     'Flames envelop you, causing <ansi fg="damage">'+String(dmgAmt)+' damage</ansi> while you writh in pain!');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is enveloped in <ansi fg="red">flames</ansi>.', actor.UserId());
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You are no longer on fire.');
    SendRoomMessage(actor.GetRoomId(),  'The healing aura surrounding '+actor.GetCharacterName(true)+' is no longer on fire.', actor.UserId());
}
