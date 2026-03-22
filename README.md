# YGOJH_Set_Probability
This repo are the tools created for the write-up [*Hot Singles in your Pack: An Analysis of Yu-Gi-Oh Card Probabilities Using Binomial Probability*](https://github.com/Lizadking/Writeups/tree/main/YGO_Justice_Hunters_Analysis)

It includes two tools:
- ygo.go 
    - This is the main tool using the [YGOPRO Api](https://ygoprodeck.com/) to get card information on the set *Justice Hunters* and calculates
    card pull probabilities using [Binomial Distribution](https://en.wikipedia.org/wiki/Binomial_distribution)
    - All data is put in the folder *data* and is organized by the following:
    ```
        .
        └── data/
            ├── Card_by_Konami_Set_Id_and_rarity/
            │   ├── card_pulls_24_packs
            │   ├── card_pulls_48_packs
            │   └── card_pulls_72_packs
            └── ...
    ```
    - build with `go build ygo.go` and run with `./ygo`

- genGraph.py
    - A simple python program to generate the graph used in the write-up's example 

** Tools are built for Linux systems only **

