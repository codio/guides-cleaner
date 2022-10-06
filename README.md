# Cleans un-used guides assets leftover from splitting books

## Usage Overview
`guides-cleaner clean-content <path_to_the_project>`

## Inline installer

```bash
sh -c "$(curl -fsSL https://raw.githubusercontent.com/codio/guides-cleaner/master/installer.sh)"
```

## How to Use Guides Cleaner in Codio
1. Download the linux binary (`content-checker-linux-amd64.tgz`) under [**Releases**](https://github.com/codio/guides-cleaner/releases)
2. Drag-and-drop `guides-cleaner-linux-amd64.tgz` into the Codio assignment or project filetree
3. In the Codio terminal, extract the binary:
    ```
    tar zxf guides-cleaner-linux-amd64.tgz
    ```
4. In the Codio terminal, run the code:
    ```
    ./guides-cleaner clean-content
    ```
## How to Merge assingments in Codio
1. Download the linux binary (`guides-cleaner-linux-amd64.tgz`) under [**Releases**](https://github.com/codio/guides-cleaner/releases)
2. Drag-and-drop `guides-cleaner-linux-amd64.tgz` into the Codio assignment or project filetree
3. In the Codio terminal, extract the binary:
    ```
    tar zxf guides-cleaner-linux-amd64.tgz
    ```
4. Clone the two assingments you want to merge into Codio
5. In the Codio terminal, run the code:
    ```
    ./guides-cleaner merge <destAssignmentPath> <mergeAssignmentPath>
    ```

    **Note:** The second assignment path (merge) is appended to the end of the first assignment path (dest)
    For example:
    ```
    ./guides-cleaner merge cs-intro-python-loops/ cs-intro-python-conditionals/
    ```
    Conditionals would be appended after the loops content

    If you want to append content into the existing content (in `.guides`), set the `<destAssignmentPath>` to `./`. Make sure your `<mergeAssignmentPath>` ends with a `/` or it will overwrite the content.
    
    To move content into `.guides` (overwriting any existing content in `.guides`), set the `<destAssignmentPath>` to `.`.

