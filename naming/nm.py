#!/usr/bin/python
normal_regexes = [
    ('standard_repeat',
     # Show.Name.S01E02.S01E03.Source.Quality.Etc-Group
     # Show Name - S01E02 - S01E03 - S01E04 - Ep Name
     '''
     ^(?P<series_name>.+?)[. _-]+                # Show_Name and separator
     s(?P<season_num>\d+)[. _-]*                 # S01 and optional separator
     e(?P<ep_num>\d+)                            # E02 and separator
     ([. _-]+s(?P=season_num)[. _-]*             # S01 and optional separator
     e(?P<extra_ep_num>\d+))+                    # E03/etc and separator
     [. _-]*((?P<extra_info>.+?)                 # Source_Quality_Etc-
     ((?<![. _-])(?<!WEB)                        # Make sure this is really the release group
     -(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$              # Group
     '''),

    ('fov_repeat',
     # Show.Name.1x02.1x03.Source.Quality.Etc-Group
     # Show Name - 1x02 - 1x03 - 1x04 - Ep Name
     '''
     ^(?P<series_name>.+?)[. _-]+                # Show_Name and separator
     (?P<season_num>\d+)x                        # 1x
     (?P<ep_num>\d+)                             # 02 and separator
     ([. _-]+(?P=season_num)x                    # 1x
     (?P<extra_ep_num>\d+))+                     # 03/etc and separator
     [. _-]*((?P<extra_info>.+?)                 # Source_Quality_Etc-
     ((?<![. _-])(?<!WEB)                        # Make sure this is really the release group
     -(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$              # Group
     '''),

    ('standard',
     # Show.Name.S01E02.Source.Quality.Etc-Group
     # Show Name - S01E02 - My Ep Name
     # Show.Name.S01.E03.My.Ep.Name
     # Show.Name.S01E02E03.Source.Quality.Etc-Group
     # Show Name - S01E02-03 - My Ep Name
     # Show.Name.S01.E02.E03
     '''
     ^((?P<series_name>.+?)[. _-]+)?             # Show_Name and separator
     (\()?s(?P<season_num>\d+)[. _-]*            # S01 and optional separator
     e(?P<ep_num>\d+)(\))?                       # E02 and separator
     (([. _-]*e|-)                               # linking e/- char
     (?P<extra_ep_num>(?!(1080|720|480)[pi])\d+)(\))?)*   # additional E03/etc
     [. _-]*((?P<extra_info>.+?)                 # Source_Quality_Etc-
     ((?<![. _-])(?<!WEB)                        # Make sure this is really the release group
     -(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$              # Group
     '''),

    ('fov',
     # Show_Name.1x02.Source_Quality_Etc-Group
     # Show Name - 1x02 - My Ep Name
     # Show_Name.1x02x03x04.Source_Quality_Etc-Group
     # Show Name - 1x02-03-04 - My Ep Name
     '''
     ^((?P<series_name>.+?)[\[. _-]+)?           # Show_Name and separator
     (?P<season_num>\d+)x                        # 1x
     (?P<ep_num>\d+)                             # 02 and separator
     (([. _-]*x|-)                               # linking x/- char
     (?P<extra_ep_num>
     (?!(1080|720|480)[pi])(?!(?<=x)264)             # ignore obviously wrong multi-eps
     \d+))*                                      # additional x03/etc
     [\]. _-]*((?P<extra_info>.+?)               # Source_Quality_Etc-
     ((?<![. _-])(?<!WEB)                        # Make sure this is really the release group
     -(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$              # Group
     '''),

    ('scene_date_format',
     # Show.Name.2010.11.23.Source.Quality.Etc-Group
     # Show Name - 2010-11-23 - Ep Name
     '''
     ^((?P<series_name>.+?)[. _-]+)?             # Show_Name and separator
     (?P<air_date>(\d+[. _-]\d+[. _-]\d+)|(\d+\w+[. _-]\w+[. _-]\d+))
     [. _-]*((?P<extra_info>.+?)                 # Source_Quality_Etc-
     ((?<![. _-])(?<!WEB)                        # Make sure this is really the release group
     -(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$              # Group
     '''),

    ('scene_sports_format',
     # Show.Name.100.Event.2010.11.23.Source.Quality.Etc-Group
     # Show.Name.2010.11.23.Source.Quality.Etc-Group
     # Show Name - 2010-11-23 - Ep Name
     '''
     ^(?P<series_name>.*?(UEFA|MLB|ESPN|WWE|MMA|UFC|TNA|EPL|NASCAR|NBA|NFL|NHL|NRL|PGA|SUPER LEAGUE|FORMULA|FIFA|NETBALL|MOTOGP).*?)[. _-]+
     ((?P<series_num>\d{1,3})[. _-]+)?
     (?P<air_date>(\d+[. _-]\d+[. _-]\d+)|(\d+\w+[. _-]\w+[. _-]\d+))[. _-]+
     ((?P<extra_info>.+?)((?<![. _-])
     (?<!WEB)-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$
     '''),

    ('stupid',
     # tpz-abc102
     '''
     (?P<release_group>.+?)-\w+?[\. ]?           # tpz-abc
     (?!264)                                     # don't count x264
     (?P<season_num>\d{1,2})                     # 1
     (?P<ep_num>\d{2})$                          # 02
     '''),

    ('verbose',
     # Show Name Season 1 Episode 2 Ep Name
     '''
     ^(?P<series_name>.+?)[. _-]+                # Show Name and separator
     season[. _-]+                               # season and separator
     (?P<season_num>\d+)[. _-]+                  # 1
     episode[. _-]+                              # episode and separator
     (?P<ep_num>\d+)[. _-]+                      # 02 and separator
     (?P<extra_info>.+)$                         # Source_Quality_Etc-
     '''),

    ('season_only',
     # Show.Name.S01.Source.Quality.Etc-Group
     '''
     ^((?P<series_name>.+?)[. _-]+)?             # Show_Name and separator
     s(eason[. _-])?                             # S01/Season 01
     (?P<season_num>\d+)[. _-]*                  # S01 and optional separator
     [. _-]*((?P<extra_info>.+?)                 # Source_Quality_Etc-
     ((?<![. _-])(?<!WEB)                        # Make sure this is really the release group
     -(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$              # Group
     '''
    ),

    ('no_season_multi_ep',
     # Show.Name.E02-03
     # Show.Name.E02.2010
     '''
     ^((?P<series_name>.+?)[. _-]+)?             # Show_Name and separator
     (e(p(isode)?)?|part|pt)[. _-]?              # e, ep, episode, or part
     (?P<ep_num>(\d+|[ivx]+))                    # first ep num
     ((([. _-]+(and|&|to)[. _-]+)|-)             # and/&/to joiner
     (?P<extra_ep_num>(?!(1080|720|480)[pi])(\d+|[ivx]+))[. _-])            # second ep num
     ([. _-]*(?P<extra_info>.+?)                 # Source_Quality_Etc-
     ((?<![. _-])(?<!WEB)                        # Make sure this is really the release group
     -(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$              # Group
     '''
    ),

    ('no_season_general',
     # Show.Name.E23.Test
     # Show.Name.Part.3.Source.Quality.Etc-Group
     # Show.Name.Part.1.and.Part.2.Blah-Group
     '''
     ^((?P<series_name>.+?)[. _-]+)?             # Show_Name and separator
     (e(p(isode)?)?|part|pt)[. _-]?              # e, ep, episode, or part
     (?P<ep_num>(\d+|([ivx]+(?=[. _-]))))                    # first ep num
     ([. _-]+((and|&|to)[. _-]+)?                # and/&/to joiner
     ((e(p(isode)?)?|part|pt)[. _-]?)           # e, ep, episode, or part
     (?P<extra_ep_num>(?!(1080|720|480)[pi])
     (\d+|([ivx]+(?=[. _-]))))[. _-])*            # second ep num
     ([. _-]*(?P<extra_info>.+?)                 # Source_Quality_Etc-
     ((?<![. _-])(?<!WEB)                        # Make sure this is really the release group
     -(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$              # Group
     '''
    ),

    ('no_season',
     # Show Name - 01 - Ep Name
     # 01 - Ep Name
     # 01 - Ep Name
     '''
     ^((?P<series_name>.+?)(?:[. _-]{2,}|[. _]))?    # Show_Name and separator
     (?P<ep_num>\d{1,3})                             # 02
     (?:-(?P<extra_ep_num>\d{1,3}))*                 # -03-04-05 etc
     \s?of?\s?\d{1,3}?                               # of joiner (with or without spaces) and series total ep
     [. _-]+((?P<extra_info>.+?)                     # Source_Quality_Etc-
     ((?<![. _-])(?<!WEB)                            # Make sure this is really the release group
     -(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$  # Group
     '''
    ),

    ('bare',
     # Show.Name.102.Source.Quality.Etc-Group
     '''
     ^(?P<series_name>.+?)[. _-]+                # Show_Name and separator
     (?P<season_num>\d{1,2})                     # 1
     (?P<ep_num>\d{2})                           # 02 and separator
     ([. _-]+(?P<extra_info>(?!\d{3}[. _-]+)[^-]+) # Source_Quality_Etc-
     (-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$                # Group
     '''),
]


import re

from regexes import anime_regexes

for reg in anime_regexes:
    news = ""
    linenum = 0
    for l in reg[1].split("\n"):
        if not l.strip():
            continue
        parts = l.split("#")
        if linenum == 0:
            parts[0] = "(?i)"+parts[0].strip()
        if len(parts) > 1:
            news += "`{0}`+ // {1}\n".format(parts[0].strip(), parts[1])
        else:
            news += "`{0}`+\n".format(l.strip())
        linenum = linenum + 1

    s = '''  {{
    Name:  "{0}",
    Regex: pcre.MustCompile({1}}},'''.format(reg[0], news)
    sl = list(s)
    sl[s.rfind("+")] = ", pcre.CASELESS),"
    s = ''.join(sl)
    print s
