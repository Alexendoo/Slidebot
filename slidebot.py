import re
import znc

class slidebot(znc.Module):
    description = "Does github stuff maybe"

    def OnChanMsg(self, nick, channel, message):
        chan = channel.GetName()
        nick = nick.GetNick()
        msg = message.s

        if ("Not-" in nick) or (nick == "gonzobot"):
            return znc.CONTINUE

        issue = re.search('(?<=#)[0-9]{1,3}', msg)

        if issue:
            self.PutIRC("PRIVMSG {0} :https://github.com/ccrama/Slide/issues/{1}".format(chan, issue.group(0)))
            return znc.CONTINUE

        # self.PutModule("Hey, {0} said {1} on {2}".format(nick, msg, chan))
        self.PutIRC("PRIVMSG {0} :{1}".format(chan, msg))
        return znc.CONTINUE
