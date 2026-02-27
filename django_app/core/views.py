from django.contrib.auth.decorators import login_required
from django.shortcuts import render


@login_required
def index(request):
    return render(request, "core/index.html", {"user": request.user})


@login_required
def user_info(request):
    return render(request, "core/user_info.html", {"user": request.user})

