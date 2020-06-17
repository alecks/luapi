import React from "react";
import MonacoEditor from "react-monaco-editor";
import starter from "./starter";

interface State {
  code: string;
}

interface LuAPIResponse {
  body: string;
  status: number;
}

class App extends React.Component<Readonly<{}>, State> {
  private static readonly runRegex = /^--!/gm;

  constructor(props: Readonly<{}>) {
    super(props);
    this.state = {
      code: starter,
    };

    this.onChange = this.onChange.bind(this);
  }

  async onChange(script: string) {
    if (script.match(App.runRegex)) {
      const [apiURL, namespace] = script
        .match(/^--:( *)?.*/gm)
        ?.map((v) => v.replace(/^--:( *)?/gm, "")) || [
        "http://luapi.example.org",
        "global",
      ];
      localStorage.setItem("defaultURL", apiURL);

      const res = (await fetch(apiURL, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          script,
          namespace,
        }),
      }).catch((e) =>
        this.setState({
          code: script.replace(
            App.runRegex,
            `--[[
  Error:
  ${e.message}
--]]`
          ),
        })
      )) as Response;

      if (res) {
        const json = (await res.json()) as LuAPIResponse;

        this.setState({
          code: script.replace(
            App.runRegex,
            `--[[
  Status: ${res.status} ${res.statusText}
  ${json.body}
--]]`
          ),
        });
      }
    }
  }

  render() {
    return (
      <MonacoEditor
        width="100vw"
        height="100vh"
        language="lua"
        theme="vs-dark"
        value={this.state.code}
        onChange={this.onChange}
      />
    );
  }
}

export default App;
