export default `--[[
  Welcome to the LuAPI editor!
  This is a simple request template; feel free to edit it.


  There's a few special comments that you can use to work your way around LuAPI.
  Note: these need to be on their own lines.

    To run/send the request:
      --!

    To set a configuration parameter:
      --: <value>
    As for configurations, the first comment is always the server URL and the second is the namespace.


  Here's some general LuAPI functions:

    respond(result string):
      This is the equivalent to print. You can only call this once; the first value is always prioritised.
--]]

-- Make sure to edit these values to fit your server.
--: http://localhost
--: global

function main() return 'Hello, world!'; end

respond(main());
`;
