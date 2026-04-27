function onStart(actor, triggersLeft) {
    actor.SetAdjective("polymorphed", true);
}

function onTrigger(actor, triggersLeft) {
    if (triggersLeft == 5) {
        SendUserMessage(actor.UserId(), 'You feel the polymorph weakening...');
    }
}

function onEnd(actor, triggersLeft) {
    actor.RevertFormChange();
    SendUserMessage(actor.UserId(), 'The polymorph fades and your body shifts back to normal.');
    SendRoomMessage(actor.GetRoomId(),
        actor.GetCharacterName(true) + ' shimmers and reverts to their true form.',
        actor.UserId());
    actor.SetAdjective("polymorphed", false);
}
