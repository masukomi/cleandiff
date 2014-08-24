#!/usr/bin/env sh


git config --global difftool.cdiff.cmd "cdiff \"\$LOCAL\" \"\$REMOTE\" \"\$MERGED\""
echo "- Registered \"cdiff\" diff tool with git."

PROMPT=$(cat <<EOL
Make CleanDiff your *default* git diffing tool?
(it is awesome...)[y|n]
EOL
)
while true; do
    read -p "$PROMPT" yn
    case $yn in
        [Yy]* ) echo "I love you too."; git config --global diff.tool cdiff; break;;
        [Nn]* ) echo "Maybe later."; break;;
        * ) echo "Please answer yes or no.";;
    esac
done

echo "Done."
