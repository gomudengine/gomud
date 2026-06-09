
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    dmgAmt = Math.abs(actor.AddHealth(-1*(UtilDiceRoll(2, 9)+2)));

    SendUserMessage(actor.UserId(),     'Fiery shrapnel hits you for <ansi fg="damage">'+String(dmgAmt)+' damage</ansi>!');
    SendRoomMessage(actor.GetRoomId(),  'Fiery shrapnel hits '+actor.GetCharacterName(true)+'', actor.UserId());
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    actor.GiveBuff(22, "explosion"); // On fire
}
