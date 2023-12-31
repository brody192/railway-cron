# Railway Cron.. But Better?

This Service will help you restart (or redeploy) your selected service on a Cron schedule, this differs from the native Cron feature in two ways:

- **The `restart` action:** Can trigger a service restart, instead of re-running the image in a new deployment.

- **The `redeploy` action:** Will trigger a full redeploy and replace any existing deployments. The native cron feature will wait until the deployment is in a "COMPLETED" state to run the next deployment.

Both these actions are equivalent to clicking the `restart` / `redeploy` options in the deployments 3-Dot menu but on the desired Cron schedule.

**Caveats:** Since this template uses one token for all the Railway API calls made, the token provided would need to have access to all the projects in the defined schedules.
eg, If you use a token for team 'A' all the projects in the schedules would need to be within team 'A'.

If you used a personal account token, all the projects in the schedules would need to be under your personal account.

It doesn't matter if you deploy this template to the team or personal account, as long as you provide the correct token.

## How to Setup

- **Get at least one Cron schedule prepared**

    - This is made easy with a [bookmarklet](https://bookmarklets.org/what-is-a-bookmarklet/). The pre-made bookmarklet that will help construct our schedule can be added to your bookmark bar via [this website](https://bookmarklets.up.railway.app/cron-config/)

    - Once the bookmarklet has been added to your bookmark bar, open your Railway project and desired service, and click the bookmarklet, a popup will ask you to input your Cron expression and action type, then you will be forwarded to a page that contains your schedule config, copy this long string and save it somewhere for later use.

        ```
        Field name   | Mandatory? | Allowed values  | Allowed special characters
        ----------   | ---------- | --------------  | --------------------------
        Minutes      | Yes        | 0-59            | * / , -
        Hours        | Yes        | 0-23            | * / , -
        Day of month | Yes        | 1-31            | * / , - ?
        Month        | Yes        | 1-12 or JAN-DEC | * / , -
        Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?
        ```
    
        Examples of some Cron Expressions can be found in the <a href="https://docs.railway.app/reference/cron-jobs#examples">Railway documentation</a>

- **Generate a Railway Account Token**

    - Go to the [Tokens page](https://railway.app/account/tokens) in your Railway Account, click the name field, and provide an accurate name for what the token will be used for, A name like "Cron Schedular" would suffice, Click "Create", then save this token for later use.
    
        **Note: This must be an Account Token, a project Token will not work**

- Deploy this template

    [![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/new/template/fwH-l3?referralCode=ySCnWl)

    - Input the schedule configuration string into the `SCHEDULE_1` variable.

        - More schedules can be added later by using the bookmarklet and by adding the resulting schedule configuration string to a `SCHEDULE_2`
 variable, and so on.

    - Input your **Account** Token into the `RAILWAY_ACCOUNT_TOKEN` variable.

    - Click "Deploy"

    - You will now be able to monitor the Cron Job's Progress in the newly deployed template's deployment logs