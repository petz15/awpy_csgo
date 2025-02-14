import sys
import os
import time
# from demoparser_V2 import DemoParser
from demoparser_V3 import DemoParser

start_time = time.time()

error_foloder = r"C:\D\coding_projects\csgo_demo_parser_python\error demos\go_test"

good_folder = r"C:\D\coding_projects\csgo_demo_parser_python\demos\2354989\2354989"

os.chdir(good_folder)


demo_parser = DemoParser(
    demofile = "ence-vs-faze-m1-vertigo.dem", 
    demo_id = "2354989_m1", 
    log=True,
    parse_frames=False ,
    json_indentation= True,
    buy_style="hltv")

    
data = demo_parser.parse()

end_time = time.time()
elapsed_time = end_time - start_time
print(f"Program runtime: {elapsed_time} seconds")

sys.exit()