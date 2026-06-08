
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You enter a focused state of rest.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' begins to meditate.', actor.UserId());
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {

    skillLevel = actor.GetSkillLevel("brawling");
    
    maxHealing = 4;
    if (skillLevel == 3) {
        maxHealing = 6;
    } else if (skillLevel >= 4) {
        maxHealing = 8;
    }

    healAmt = actor.AddHealth(UtilDiceRoll(1, maxHealing));

    SendUserMessage(actor.UserId(),     'You heal for <ansi fg="healing">'+String(healAmt)+' hitpoints</ansi>.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is healing while they meditate.', actor.UserId());
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'Your restful state abides.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is done meditating.', actor.UserId());
}
