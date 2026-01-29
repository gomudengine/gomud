# Help for ~picklock~

The ~picklock~ command attempts to pick the lock of any exit or container.  
You'll need a set of *lockpicks*.

## Usage:

  ~picklock [exit name/container name]~ - This picks the lock. You can also use it to inspect the lock and select `quit` without attempting to pick it.

## Examples:

    **[HP:66/66 MP:35/35]:** ~north~  
    There's a lock preventing you from going that way. You'll need a *Key* or to ~pick~ the lock with *lockpicks*.


    **[HP:66/66 MP:35/35]:** ~pick north~

    **The Lock Sequence Looks like:**  
    **╔═══════╦═══════╦═══════╦═══════╦═══════╗**  
    **║       ║       ║       ║       ║       ║**  
    **║   ?   ║   ?   ║   ?   ║   ?   ║   ?   ║**  
    **║       ║       ║       ║       ║       ║**  
    **╚═══════╩═══════╩═══════╩═══════╩═══════╝**  
    **Move your lockpick?** **[**UP**/**DOWN**/**quit**]**

**Note:** *Each **?** symbol represents an UP or DOWN you must guess in the sequence.*

**See also:** ~help picklock-example~, ~help keyring~