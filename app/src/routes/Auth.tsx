import { tokenField } from "@/conf/bootstrap.ts";
import { useEffect, useReducer } from "react";
import Loader from "@/components/Loader.tsx";
import "@/assets/pages/auth.less";
import { validateToken } from "@/store/auth.ts";
import { useDispatch } from "react-redux";
import router from "@/router.tsx";
import { useTranslation } from "react-i18next";
import { getQueryParam } from "@/utils/path.ts";
import { setMemory } from "@/utils/memory.ts";
import { appLogo, appName, useDeeptrain } from "@/conf/env.ts";
import { Card, CardContent } from "@/components/ui/card.tsx";
import { goAuth } from "@/utils/app.ts";
import { Label } from "@/components/ui/label.tsx";
import { Input } from "@/components/ui/input.tsx";
import Require, { LengthRangeRequired } from "@/components/Require.tsx";
import { Button } from "@/components/ui/button.tsx";
import { formReducer, isTextInRange } from "@/utils/form.ts";
import { doLogin, doWechatLogin } from "@/api/auth.ts";
import { getErrorMessage, isEnter } from "@/utils/base.ts";
import { ScrollArea } from "@/components/ui/scroll-area.tsx";
import { toast } from "sonner";

function DeepAuth() {
  const { t } = useTranslation();
  const dispatch = useDispatch();
  const token = getQueryParam("token").trim();

  useEffect(() => {
    if (!token.length) {
      toast.warning(t("invalid-token"), {
        description: t("invalid-token-prompt"),
        action: {
          label: t("try-again"),
          onClick: goAuth,
        },
      });

      setTimeout(goAuth, 2500);
      return;
    }

    setMemory(tokenField, token);

    doLogin({ token })
      .then((data) => {
        if (!data.status) {
          toast.error(t("login-failed"), {
            description: t("login-failed-prompt", { reason: data.error }),
            action: {
              label: t("try-again"),
              onClick: goAuth,
            },
          });
        } else
          validateToken(dispatch, data.token, async () => {
            toast.success(t("login-success"), {
              description: t("login-success-prompt"),
            });

            await router.navigate("/");
          });
      })
      .catch((err) => {
        console.debug(err);

        toast.error(t("server-error"), {
          description: `${t("server-error-prompt")}\n${err.message}`,
          action: {
            label: t("try-again"),
            onClick: goAuth,
          },
        });
      });
  }, []);

  return (
    <div className={`auth`}>
      <Loader prompt={t("login")} />
    </div>
  );
}

function Login() {
  const { t } = useTranslation();
  const globalDispatch = useDispatch();
  const [, dispatch] = useReducer(formReducer<Record<string, string>>(), {});

  const onSubmit = async () => {
    // TODO: 接入微信网页/PC OpenSDK，这里从 SDK 获得 code 后再调用登录。
    const codeFromSdk = (window as any)?.wechatLoginCode || getQueryParam("code") || "";

    if (!isTextInRange(codeFromSdk, 1, 255)) {
      toast.warning("请先完成微信授权", {
        description: "接入微信 SDK 获取 code 后再点击快捷登录。",
      });
      return;
    }

    try {
      const resp = await doWechatLogin({ code: codeFromSdk });
      if (!resp.status) {
        toast.warning(t("login-failed"), {
          description: t("login-failed-prompt", { reason: resp.error }),
        });
        return;
      }

      toast.success(t("login-success"), {
        description: t("login-success-prompt"),
      });

      validateToken(globalDispatch, resp.token);
      await router.navigate("/");
    } catch (err) {
      console.debug(err);
      toast.error(t("server-error"), {
        description: `${t("server-error-prompt")}\n${getErrorMessage(err)}`,
      });
    }
  };

  useEffect(() => {
    // listen to enter key and auto submit
    const listener = async (e: KeyboardEvent) => {
      if (isEnter(e)) await onSubmit();
    };

    document.addEventListener("keydown", listener);
    return () => document.removeEventListener("keydown", listener);
  }, []);

  return (
    <ScrollArea className={`w-full h-full grid place-items-center`}>
      <div className={`auth-container`}>
        <img className={`logo`} src={appLogo} alt="" />
        <div className={`title`}>
          {t("login")} {appName}
        </div>
        <Card className={`auth-card`}>
          <CardContent className={`pb-0`}>
            <div className={`auth-wrapper`}>
              <Button
                tapScale={0.975}
                classNameWrapper={`mt-3 w-full`}
                onClick={onSubmit}
                className={`w-full`}
                loading={true}
              >
                微信快捷登录
              </Button>

              <div className={`text-sm text-muted-foreground mt-3 mb-2`}>
                使用微信扫码/授权后，前端 SDK 获取到 code，将自动完成登录。
                <br />
                TODO: 接入微信网页/PC OpenSDK，点击上方按钮后拉起微信授权。
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </ScrollArea>
  );
}

function Auth() {
  return useDeeptrain ? <DeepAuth /> : <Login />;
}

export default Auth;
