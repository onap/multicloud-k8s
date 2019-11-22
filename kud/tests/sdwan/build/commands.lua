-- Licensed to the public under the GNU General Public License v2.

module("luci.controller.commands", package.seeall)

sys = require "luci.sys"
ut = require "luci.util"
io = require "io"

ip = "ip -4 "

function index()
    entry({"admin", "config", "command"},
	call("execute")).dependent = false
end

function trim(s)
    return s:match("^%s*(.-)%s*$")
end

function split_and_trim(str, sep)
    local array = {}
    local reg = string.format("([^%s]+)", sep)
    for item in string.gmatch(str, reg) do
        item_trimed = trim(item)
        if string.len(item_trimed) > 0 then
            table.insert(array, item_trimed)
        end
    end
    return array
end

function execute()
    local commands = luci.http.formvalue("command")
    io.stderr:write("Execute command: %s\n" % commands)

    local command_array = split_and_trim(commands, ";")
    for index, command in ipairs(command_array) do
        sys.exec(command)
    end

    luci.http.prepare_content("application/json")
    luci.http.write_json("{'status':'ok'}")
end
