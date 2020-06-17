--[[
    bootstrapper.lua is responsible for bootstrapping all incoming scripts.
    This should include — for example — the creation a sandbox, and possibly global functions/variables.
    To create namespace-specific funcs/vars, all you have to do is run DoString/File twice.
--]]

print("Bootstrapping LuAPI!");
debug = nil; io = nil; os = nil