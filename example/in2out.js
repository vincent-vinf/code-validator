const readline = require('readline').createInterface({
    input: process.stdin,
    output: process.stdout
});

readline.question('', response => {
    console.log(response)
    readline.close();
});
