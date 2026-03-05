// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React from "react";
import {Modal} from "antd";
import i18next from "i18next";
import * as Provider from "./Provider";
import {getProviderLogoURL} from "../Setting";
import {GithubLoginButton, GoogleLoginButton} from "react-social-login-buttons";
import QqLoginButton from "./QqLoginButton";
import FacebookLoginButton from "./FacebookLoginButton";
import WeiboLoginButton from "./WeiboLoginButton";
import GiteeLoginButton from "./GiteeLoginButton";
import WechatLoginButton from "./WechatLoginButton";
import DingTalkLoginButton from "./DingTalkLoginButton";
import LinkedInLoginButton from "./LinkedInLoginButton";
import WeComLoginButton from "./WeComLoginButton";
import LarkLoginButton from "./LarkLoginButton";
import GitLabLoginButton from "./GitLabLoginButton";
import AdfsLoginButton from "./AdfsLoginButton";
import CasdoorLoginButton from "./CasdoorLoginButton";
import BaiduLoginButton from "./BaiduLoginButton";
import AlipayLoginButton from "./AlipayLoginButton";
import InfoflowLoginButton from "./InfoflowLoginButton";
import AppleLoginButton from "./AppleLoginButton";
import AzureADLoginButton from "./AzureADLoginButton";
import AzureADB2CLoginButton from "./AzureADB2CLoginButton";
import SlackLoginButton from "./SlackLoginButton";
import SteamLoginButton from "./SteamLoginButton";
import BilibiliLoginButton from "./BilibiliLoginButton";
import OktaLoginButton from "./OktaLoginButton";
import DouyinLoginButton from "./DouyinLoginButton";
import KwaiLoginButton from "./KwaiLoginButton";
import LoginButton from "./LoginButton";
import * as AuthBackend from "./AuthBackend";
import {WechatOfficialAccountModal} from "./Util";
import * as Setting from "../Setting";
import IdtrustImg from "../static/idtrust.png";

function getSigninButton(provider) {
  const text = i18next.t("login:Sign in with {type}").replace("{type}", provider.displayName !== "" ? provider.displayName : provider.type);
  if (provider.type === "GitHub") {
    return <GithubLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Google") {
    return <GoogleLoginButton text={text} align={"center"} />;
  } else if (provider.type === "QQ") {
    return <QqLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Facebook") {
    return <FacebookLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Weibo") {
    return <WeiboLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Gitee") {
    return <GiteeLoginButton text={text} align={"center"} />;
  } else if (provider.type === "WeChat") {
    return <WechatLoginButton text={text} align={"center"} />;
  } else if (provider.type === "DingTalk") {
    return <DingTalkLoginButton text={text} align={"center"} />;
  } else if (provider.type === "LinkedIn") {
    return <LinkedInLoginButton text={text} align={"center"} />;
  } else if (provider.type === "WeCom") {
    return <WeComLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Lark") {
    return <LarkLoginButton text={text} align={"center"} />;
  } else if (provider.type === "GitLab") {
    return <GitLabLoginButton text={text} align={"center"} />;
  } else if (provider.type === "ADFS") {
    return <AdfsLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Casdoor") {
    return <CasdoorLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Baidu") {
    return <BaiduLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Alipay") {
    return <AlipayLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Infoflow") {
    return <InfoflowLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Apple") {
    return <AppleLoginButton text={text} align={"center"} />;
  } else if (provider.type === "AzureAD") {
    return <AzureADLoginButton text={text} align={"center"} />;
  } else if (provider.type === "AzureADB2C") {
    return <AzureADB2CLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Slack") {
    return <SlackLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Steam") {
    return <SteamLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Bilibili") {
    return <BilibiliLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Okta") {
    return <OktaLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Douyin") {
    return <DouyinLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Kwai") {
    return <KwaiLoginButton text={text} align={"center"} />;
  } else {
    return <LoginButton key={provider.type} type={provider.type} logoUrl={getProviderLogoURL(provider)} />;
  }
}

function getAuthUrl(application, provider) {
  return Provider.getAuthUrl(application, provider, "signup");
}

function handleInviterCodeState(inviterCode, isFromWeb, invitationChecked) {
  // 获取当前URL参数
  const url = new URL(window.location.href);
  const params = new URLSearchParams(url.search);
  let customState = null;
  if (invitationChecked) {
    if (isFromWeb) {
      // 如果inviterCode存在，当isFromWeb为true时，替换url中的state参数为inviterCode
      customState = inviterCode;
    } else {
      // 如果inviterCode存在，当isFromWeb为false时，在原state的基础上进行拼接，增加`${原state}__inviter_code=${inviterCode}`
      const originalState = params.get("state") || "";
      // 检查原state中是否已经包含__inviter_code=，如果包含则先移除旧的，避免重复添加
      const inviterCodePattern = /__inviter_code=[^&]*/;
      let cleanState = originalState;
      if (inviterCodePattern.test(originalState)) {
        cleanState = originalState.replace(inviterCodePattern, "");
      }
      customState = `${cleanState}__inviter_code=${inviterCode}`;
    }
  } else {
    if (isFromWeb) {
      // 如果inviterCode不存在，当isFromWeb为true时，清空url中的state参数
      customState = "";
      params.delete("state");
    }
    // 如果inviterCode不存在，当isFromWeb为false时，不处理，保持原本的state参数
  }
  // 如果有自定义state，先设置到URL参数中
  if (customState !== null) {
    params.set("state", customState);
    // 更新当前页面的URL参数（不刷新页面）
    const newUrl = `${url.pathname}?${params.toString()}${url.hash}`;
    window.history.replaceState({}, "", newUrl);
  }
}

function goToSamlUrl(provider, location) {
  const params = new URLSearchParams(location.search);
  const clientId = params.get("client_id") ?? "";
  const state = params.get("state");
  const realRedirectUri = params.get("redirect_uri");
  const redirectUri = `${window.location.origin}/callback/saml`;
  const providerName = provider.name;

  const relayState = `${clientId}&${state}&${providerName}&${realRedirectUri}&${redirectUri}`;
  AuthBackend.getSamlLogin(`${provider.owner}/${providerName}`, btoa(relayState)).then((res) => {
    if (res.status === "ok") {
      if (res.data2 === "POST") {
        document.write(res.data);
      } else {
        window.location.href = res.data;
      }
    } else {
      Setting.showMessage("error", res.msg);
    }
  });
}

export function goToWeb3Url(application, provider, method) {
  if (provider.type === "MetaMask") {
    import("./Web3Auth")
      .then(module => {
        const authViaMetaMask = module.authViaMetaMask;
        authViaMetaMask(application, provider, method);
      });
  } else if (provider.type === "Web3Onboard") {
    import("./Web3Auth")
      .then(module => {
        const authViaWeb3Onboard = module.authViaWeb3Onboard;
        authViaWeb3Onboard(application, provider, method);
      });
  }
}

function getHasIdtrustProviderItems(provider) {
  return !!(provider?.name?.toLowerCase() === "idtrust"
    && provider?.type === "Custom");
}

export function renderProviderLogo(provider, application, width, margin, size, location, bindType, isFromWeb = false, inviterCode = "", invitationChecked = false, providerBtnCheckInvitationCode, onTermsChange, termsAccepted) {
  if (size === "small") {
    if (provider.category === "OAuth") {
      if (provider.type === "WeChat" && provider.clientId2 !== "" && provider.clientSecret2 !== "" && provider.disableSsl === true && !navigator.userAgent.includes("MicroMessenger")) {
        return (
          <a key={provider.displayName} >
            <img width={width} height={width} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={{margin: margin}} onClick={() => {
              WechatOfficialAccountModal(application, provider, "signup");
            }} />
          </a>
        );
      } else {
        const hasIdtrust = getHasIdtrustProviderItems(provider);
        return (
          <div style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            borderWidth: "1px",
            borderStyle: "solid",
            borderColor: "#d9d9d9",
            borderRadius: "2px",
            cursor: "pointer",
          }}
          onClick={async() => {
            // 检查 termsAccepted 状态，如果为 false 则显示提示信息
            if (!termsAccepted) {
              Modal.confirm({
                title: i18next.t("protocolModal:title"),
                content: i18next.t("protocolModal:content"),
                okText: i18next.t("protocolModal:okText"),
                cancelText: i18next.t("protocolModal:cancelText"),
                icon: null,
                centered: true,
                onOk: async() => {
                  // 点击确认后，触发勾选函数，修改勾选状态
                  if (onTermsChange) {
                    onTermsChange({target: {checked: true}});
                  }
                  // 绑定不进行处理
                  if (bindType) {
                    const authUrl = getAuthUrl(application, provider, "signup");
                    window.location.href = authUrl;
                    return;
                  }
                  // 勾选邀请人推荐先校验
                  if (invitationChecked) {
                    const res = await providerBtnCheckInvitationCode?.();
                    if (!res) {
                      return;
                    }
                  }
                  // 处理inviterCode和state参数逻辑
                  handleInviterCodeState(inviterCode, isFromWeb, invitationChecked);
                  // 最后才执行原来的跳转逻辑
                  const authUrl = getAuthUrl(application, provider, "signup");
                  window.location.href = authUrl;
                },
                onCancel: () => {
                  return;
                },
              });
            } else {
              // 如果已同意协议，直接执行原来的逻辑
              // 绑定不进行处理
              if (bindType) {
                const authUrl = getAuthUrl(application, provider, "signup");
                window.location.href = authUrl;
                return;
              }
              // 勾选邀请人推荐先校验
              if (invitationChecked) {
                const res = await providerBtnCheckInvitationCode?.();
                if (!res) {
                  return;
                }
              }
              // 处理inviterCode和state参数逻辑
              handleInviterCodeState(inviterCode, isFromWeb, invitationChecked);
              // 最后才执行原来的跳转逻辑
              const authUrl = getAuthUrl(application, provider, "signup");
              window.location.href = authUrl;
            }
          }}
          >
            <a key={provider.displayName} href={getAuthUrl(application, provider, "signup")}>
              <img width={width} height={width} src={hasIdtrust ? IdtrustImg : getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={{margin: margin}} />
            </a>
            <span style={{
              marginLeft: "6px",
              fontSize: "16px",
              fontWeight: 600,
            }}>
              {bindType
                ? i18next.t("login:bind with type").replace("{type}", provider.name)
                : i18next.t("login:Log in with type").replace("{type}", provider.name)}
            </span>
          </div>
        );
      }
    } else if (provider.category === "SAML") {
      return (
        <a key={provider.displayName} onClick={() => goToSamlUrl(provider, location)}>
          <img width={width} height={width} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={{margin: margin}} />
        </a>
      );
    } else if (provider.category === "Web3") {
      return (
        <a key={provider.displayName} onClick={() => goToWeb3Url(application, provider, "signup")}>
          <img width={width} height={width} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={{margin: margin}} />
        </a>
      );
    }
  } else if (provider.type === "Custom") {
    // style definition
    const text = i18next.t("login:Sign in with {type}").replace("{type}", provider.displayName);
    const customAStyle = {display: "block", height: "55px", color: "#000"};
    const customButtonStyle = {display: "flex", alignItems: "center", width: "calc(100% - 10px)", height: "50px", margin: "5px", padding: "0 10px", backgroundColor: "transparent", boxShadow: "0px 1px 3px rgba(0,0,0,0.5)", border: "0px", borderRadius: "3px", cursor: "pointer"};
    const customImgStyle = {justfyContent: "space-between"};
    const customSpanStyle = {textAlign: "center", width: "100%", fontSize: "19px"};
    if (provider.category === "OAuth") {
      return (
        <a key={provider.displayName} href={Provider.getAuthUrl(application, provider, "signup")} style={customAStyle}>
          <div style={customButtonStyle}>
            <img width={26} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={customImgStyle} />
            <span style={customSpanStyle}>{text}</span>
          </div>
        </a>
      );
    } else if (provider.category === "SAML") {
      return (
        <a key={provider.displayName} onClick={() => goToSamlUrl(provider, location)} style={customAStyle}>
          <div style={customButtonStyle}>
            <img width={26} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={customImgStyle} />
            <span style={customSpanStyle}>{text}</span>
          </div>
        </a>
      );
    }
  } else {
    // big button, for disable password signin
    if (provider.category === "SAML") {
      return (
        <div key={provider.displayName} className="provider-big-img">
          <a onClick={() => goToSamlUrl(provider, location)}>
            {
              getSigninButton(provider)
            }
          </a>
        </div>
      );
    } else if (provider.category === "Web3") {
      return (
        <div key={provider.displayName} className="provider-big-img">
          <a onClick={() => goToWeb3Url(application, provider, "signup")}>
            {
              getSigninButton(provider)
            }
          </a>
        </div>
      );
    } else {
      return (
        <div key={provider.displayName} className="provider-big-img">
          <a href={Provider.getAuthUrl(application, provider, "signup")}>
            {
              getSigninButton(provider)
            }
          </a>
        </div>
      );
    }
  }
}
