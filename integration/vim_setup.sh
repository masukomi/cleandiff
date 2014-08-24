#!/usr/bin/env sh

echo "au! BufRead,BufNewFile *.cdf  setfiletype cdiff" >> ~/.vim/filetype.vim

echo "au BufNewFile,BufRead *.cdiff set filetype=cdiff" >>  ~/.vim/ftdetect/cdiff.vim
echo "au BufNewFile,BufRead *.cdf set filetype=cdiff" \
>> ~/.vim/ftdetect/cdiff.vim
