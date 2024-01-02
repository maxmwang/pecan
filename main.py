import discord
import docker

class Cean(discord.Client):
    async def on_ready(self):
        print('Logged on as {0}!'.format(self.user))

    async def on_message(self, message):
        # don't respond to ourselves
        if message.author == self.user:
            return

        if message.content.startswith('>exec'):
            # TODO: add support for specifying language
            code = message.content.split('```')[1]

            client = docker.from_env()

            # TODO: sanitize input
            container = client.containers.run('python', 'python -c \"' + code + '\"', detach=True)

            print(container.logs())
            await message.channel.send(container.logs().decode('utf-8'))


        if message.content == 'ping':
            await message.channel.send('pong')

if __name__ == '__main__':
    intents = discord.Intents.default()
    intents.message_content = True
    client = Cean(intents=intents)
    client.run('~')

